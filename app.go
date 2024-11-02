package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/tedacmc/tedac/tedac"
	"github.com/tedacmc/tedac/tedac/chunk"
	"github.com/tedacmc/tedac/tedac/latestmappings"
	"github.com/tedacmc/tedac/tedac/legacyprotocol/legacypacket"
	"github.com/wailsapp/wails/lib/renderer/webview"
	"golang.org/x/oauth2"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// App ...
type App struct {
	listener      *minecraft.Listener
	remoteAddress string
	localPort     uint16

	src oauth2.TokenSource
	ctx context.Context

	c chan interface{}
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{src: tokenSource(), c: make(chan interface{})}
}

// ProxyInfo ...
type ProxyInfo struct {
	RemoteAddress string `json:"remote_address"`
	LocalAddress  string `json:"local_address"`
}

// ProxyingInfo returns info about the current Tedac connection. If no connection is active, an error is returned.
func (a *App) ProxyingInfo() (ProxyInfo, error) {
	if a.listener == nil {
		return ProxyInfo{}, errors.New("no connection active")
	}
	return ProxyInfo{
		RemoteAddress: a.remoteAddress,
		LocalAddress:  fmt.Sprintf("127.0.0.1:%d", a.localPort),
	}, nil
}

// Terminate terminates any existing Tedac connection.
func (a *App) Terminate() {
	if a.listener == nil {
		return
	}
	a.c <- struct{}{}
	_ = a.listener.Close()
}

// Connect starts Tedac and connects to a remote server.
func (a *App) Connect(address string) error {
	temp, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}
	l, err := net.ListenUDP("udp", temp)
	if err != nil {
		return err
	}

	port := l.LocalAddr().(*net.UDPAddr).Port
	if err = l.Close(); err != nil {
		return err
	}

	p, err := minecraft.NewForeignStatusProvider(address)
	if err != nil {
		return err
	}

	err = os.Mkdir("packcache", 0644)
	useCache := err == nil || os.IsExist(err)

	var cachedPackNames []string
	conn, err := minecraft.Dialer{
		TokenSource: a.src,
		DownloadResourcePack: func(id uuid.UUID, version string, _, _ int) bool {
			if useCache {
				name := fmt.Sprintf("%s_%s", id, version)
				_, err = os.Stat(fmt.Sprintf("packcache/%s.mcpack", name))
				if err == nil {
					cachedPackNames = append(cachedPackNames, name)
					return false
				}
			}
			return true
		},
	}.Dial("raknet", address)
	if err != nil {
		return err
	}
	packs := conn.ResourcePacks()
	_ = conn.Close()

	var cachedPacks []*resource.Pack
	if useCache {
		for _, name := range cachedPackNames {
			pack, err := resource.ReadPath(fmt.Sprintf("packcache/%s.mcpack", name))
			if err != nil {
				continue
			}
			cachedPacks = append(cachedPacks, pack)
		}
		for _, pack := range packs {
			packData := make([]byte, pack.Len())
			_, err = pack.ReadAt(packData, 0)
			if err != nil {
				continue
			}
			name := fmt.Sprintf("%s_%s", pack.UUID(), pack.Version())
			_ = os.WriteFile(fmt.Sprintf("packcache/%s.mcpack", name), packData, 0644)
		}
	}

	a.remoteAddress = address
	a.localPort = uint16(port)

	go a.startRPC()

	a.listener, err = minecraft.ListenConfig{
		AllowInvalidPackets: true,
		AllowUnknownPackets: true,

		StatusProvider:    p,
		ResourcePacks:     append(packs, cachedPacks...),
		AcceptedProtocols: []minecraft.Protocol{tedac.Protocol{}},
	}.Listen("raknet", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	go func() {
		for {
			c, err := a.listener.Accept()
			if err != nil {
				break
			}
			go a.handleConn(c.(*minecraft.Conn))
		}
	}()
	return nil
}

// CheckNetIsolation checks if a loopback exempt is in place to allow the hosting device to join the server. This is
// only relevant on Windows.
func (a *App) CheckNetIsolation() bool {
	if runtime.GOOS != "windows" {
		// Only an issue on Windows.
		return true
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
	return bytes.Contains(data, []byte("microsoft.minecraftuwp_8wekyb3d8bbwe"))
}

// startup is called when the app starts. The context is saved, so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

var (
	// airRID is the runtime ID of the air block in the latest version of the game.
	airRID, _ = latestmappings.StateToRuntimeID("minecraft:air", nil)
	// defaultSkinResourcePatch holds the skin resource patch assigned to a player when they wear a custom skin.
	defaultSkinResourcePatch = base64.StdEncoding.EncodeToString([]byte(`
		{
		   "geometry" : {
		      "default" : "geometry.humanoid.custom"
		   }
		}
	`))
)

// handleConn handles a new incoming minecraft.Conn from the minecraft.Listener passed.
func (a *App) handleConn(conn *minecraft.Conn) {
	clientData := conn.ClientData()
	if _, ok := conn.Protocol().(tedac.Protocol); ok { // TODO: Adjust this inside Protocol itself.
		clientData.GameVersion = protocol.CurrentVersion
		clientData.SkinResourcePatch = defaultSkinResourcePatch
		clientData.DeviceModel = "TEDAC CLIENT"

		data, _ := base64.StdEncoding.DecodeString(clientData.SkinData)
		switch len(data) {
		case 32 * 64 * 4:
			clientData.SkinImageHeight = 32
			clientData.SkinImageWidth = 64
		case 64 * 64 * 4:
			clientData.SkinImageHeight = 64
			clientData.SkinImageWidth = 64
		case 128 * 128 * 4:
			clientData.SkinImageHeight = 128
			clientData.SkinImageWidth = 128
		}
	}

	serverConn, err := minecraft.Dialer{
		TokenSource: a.src,
		ClientData:  clientData,
	}.Dial("raknet", a.remoteAddress)
	if err != nil {
		panic(err)
	}

	data := serverConn.GameData()

	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(data); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()

	// TODO: Component-ize the shit below.
	rid := data.EntityRuntimeID
	oldMovementSystem := data.PlayerMovementSettings.MovementType == protocol.PlayerMovementModeClient
	if _, ok := conn.Protocol().(tedac.Protocol); ok {
		oldMovementSystem = true
	}

	r := world.Overworld.Range()
	pos := atomic.NewValue(data.PlayerPosition)
	lastPos := atomic.NewValue(data.PlayerPosition)
	yaw, pitch := atomic.NewValue(data.Yaw), atomic.NewValue(data.Pitch)

	startedSneaking, stoppedSneaking := atomic.NewValue(false), atomic.NewValue(false)
	startedSprinting, stoppedSprinting := atomic.NewValue(false), atomic.NewValue(false)
	startedGliding, stoppedGliding := atomic.NewValue(false), atomic.NewValue(false)
	startedSwimming, stoppedSwimming := atomic.NewValue(false), atomic.NewValue(false)
	startedJumping := atomic.NewValue(false)

	biomeBufferCache := make(map[protocol.ChunkPos][]byte)

	if oldMovementSystem {
		go func() {
			t := time.NewTicker(time.Second / 20)
			defer t.Stop()

			var tick uint64
			for range t.C {
				currentPos, originalPos := pos.Load(), lastPos.Load()
				lastPos.Store(currentPos)

				currentYaw, currentPitch := yaw.Load(), pitch.Load()

				inputs := uint64(0)
				if startedSneaking.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStartSneaking
				}
				if stoppedSneaking.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStopSneaking
				}
				if startedSprinting.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStartSprinting
				}
				if stoppedSprinting.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStopSprinting
				}
				if startedGliding.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStartGliding
				}
				if stoppedGliding.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStopGliding
				}
				if startedSwimming.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStartSwimming
				}
				if stoppedSwimming.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagStopSwimming
				}
				if startedJumping.CompareAndSwap(true, false) {
					inputs |= packet.InputFlagJumping
				}

				err := serverConn.WritePacket(&packet.PlayerAuthInput{
					Delta:            currentPos.Sub(originalPos),
					HeadYaw:          currentYaw,
					InputData:        inputs,
					InputMode:        packet.InputModeMouse,
					InteractionModel: packet.InteractionModelCrosshair,
					Pitch:            currentPitch,
					PlayMode:         packet.PlayModeNormal,
					Position:         currentPos,
					Tick:             tick,
					Yaw:              currentYaw,
				})
				if err != nil {
					return
				}
				_ = serverConn.Flush()
				tick++
			}
		}()
	}
	go func() {
		defer a.listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			switch pk := pk.(type) {
			case *packet.MovePlayer:
				if !oldMovementSystem {
					break
				}
				pos.Store(pk.Position)
				yaw.Store(pk.Yaw)
				pitch.Store(pk.Pitch)
				continue
			case *packet.PlayerAction:
				if !oldMovementSystem {
					break
				}
				switch pk.ActionType {
				case legacypacket.PlayerActionJump:
					startedJumping.Store(true)
					continue
				case legacypacket.PlayerActionStartSprint:
					startedSprinting.Store(true)
					continue
				case legacypacket.PlayerActionStopSprint:
					stoppedSprinting.Store(true)
					continue
				case legacypacket.PlayerActionStartSneak:
					startedSneaking.Store(true)
					continue
				case legacypacket.PlayerActionStopSneak:
					stoppedSneaking.Store(true)
					continue
				case legacypacket.PlayerActionStartSwimming:
					startedSwimming.Store(true)
					continue
				case legacypacket.PlayerActionStopSwimming:
					stoppedSwimming.Store(true)
					continue
				case legacypacket.PlayerActionStartGlide:
					startedGliding.Store(true)
					continue
				case legacypacket.PlayerActionStopGlide:
					stoppedGliding.Store(true)
					continue
				}
			}
			if err := serverConn.WritePacket(pk); err != nil {
				var disconnect minecraft.DisconnectError
				if errors.As(errors.Unwrap(err), &disconnect) {
					_ = a.listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
			_ = serverConn.Flush()
		}
	}()
	go func() {
		defer serverConn.Close()
		defer a.listener.Disconnect(conn, "connection lost")
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				var disconnect minecraft.DisconnectError
				if errors.As(errors.Unwrap(err), &disconnect) {
					_ = a.listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
			switch pk := pk.(type) {
			case *packet.MovePlayer:
				if !oldMovementSystem {
					break
				}
				if pk.EntityRuntimeID == rid {
					pos.Store(pk.Position)
					yaw.Store(pk.Yaw)
					pitch.Store(pk.Pitch)
				}
			case *packet.MoveActorAbsolute:
				if !oldMovementSystem {
					break
				}
				if pk.EntityRuntimeID == rid {
					pos.Store(pk.Position)
					yaw.Store(pk.Rotation[2])
					pitch.Store(pk.Rotation[0])
				}
			case *packet.MoveActorDelta:
				if !oldMovementSystem {
					break
				}
				if pk.EntityRuntimeID == rid {
					pos.Store(pk.Position)
					yaw.Store(pk.Rotation[2])
					pitch.Store(pk.Rotation[0])
				}
			case *packet.SubChunk:
				if _, ok := conn.Protocol().(tedac.Protocol); !ok {
					// Only Tedac clients should receive the old format.
					break
				}

				chunkBuf := bytes.NewBuffer(nil)
				blockEntities := make([]map[string]any, 0)
				for _, entry := range pk.SubChunkEntries {
					if entry.Result != protocol.SubChunkResultSuccess {
						chunkBuf.Write([]byte{
							chunk.SubChunkVersion,
							0, // The client will treat this as all air.
							uint8(entry.Offset[1]),
						})
						continue
					}

					var ind uint8
					readBuf := bytes.NewBuffer(entry.RawPayload)
					sub, err := chunk.DecodeSubChunk(airRID, r, readBuf, &ind, chunk.NetworkEncoding)
					if err != nil {
						fmt.Println(err)
						continue
					}

					var blockEntity map[string]any
					dec := nbt.NewDecoderWithEncoding(readBuf, nbt.NetworkLittleEndian)
					for {
						if err := dec.Decode(&blockEntity); err != nil {
							break
						}
						blockEntities = append(blockEntities, blockEntity)
					}

					chunkBuf.Write(chunk.EncodeSubChunk(sub, chunk.NetworkEncoding, r, int(ind)))
				}

				chunkPos := protocol.ChunkPos{pk.Position.X(), pk.Position.Z()}
				_, _ = chunkBuf.Write(append(biomeBufferCache[chunkPos], 0))
				delete(biomeBufferCache, chunkPos)

				enc := nbt.NewEncoderWithEncoding(chunkBuf, nbt.NetworkLittleEndian)
				for _, b := range blockEntities {
					_ = enc.Encode(b)
				}

				_ = conn.WritePacket(&packet.LevelChunk{
					Position:      chunkPos,
					SubChunkCount: uint32(len(pk.SubChunkEntries)),
					RawPayload:    append([]byte(nil), chunkBuf.Bytes()...),
				})
				_ = conn.Flush()
				continue
			case *packet.LevelChunk:
				if pk.SubChunkCount != protocol.SubChunkRequestModeLimitless && pk.SubChunkCount != protocol.SubChunkRequestModeLimited {
					// No changes to be made here.
					break
				}

				if _, ok := conn.Protocol().(tedac.Protocol); !ok {
					// Only Tedac clients should receive the old format.
					break
				}

				max := r.Height() >> 4
				if pk.SubChunkCount == protocol.SubChunkRequestModeLimited {
					max = int(pk.HighestSubChunk)
				}

				offsets := make([]protocol.SubChunkOffset, 0, max)
				for i := 0; i < max; i++ {
					offsets = append(offsets, protocol.SubChunkOffset{0, int8(i + (r[0] >> 4)), 0})
				}

				biomeBufferCache[pk.Position] = pk.RawPayload[:len(pk.RawPayload)-1]
				_ = serverConn.WritePacket(&packet.SubChunkRequest{
					Position: protocol.SubChunkPos{pk.Position.X(), 0, pk.Position.Z()},
					Offsets:  offsets,
				})
				_ = serverConn.Flush()
				continue
			case *packet.Transfer:
				a.remoteAddress = fmt.Sprintf("%s:%d", pk.Address, pk.Port)

				pk.Address = "127.0.0.1"
				pk.Port = a.localPort
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
			_ = conn.Flush()
		}
	}()
}

// tokenSource returns a token source for using with a gophertunnel client. It either reads it from the
// token.tok file if cached or requests logging in with a device code.
func tokenSource() oauth2.TokenSource {
	token := new(oauth2.Token)
	tokenData, err := os.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token = requestToken()
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		src = auth.RefreshTokenSource(requestToken())
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile("token.tok", b, 0644)
	return src
}

// requestToken opens a new WebView2 window and requests the user to log in. The token is returned if successful.
func requestToken() *oauth2.Token {
	resp, err := auth.StartDeviceAuth()
	if err != nil {
		panic(err)
	}
	view := webview.NewWebview(webview.Settings{
		Title:  "Tedac Authentication",
		URL:    "https://login.live.com/oauth20_remoteconnect.srf?lc=1033&otc=" + resp.UserCode,
		Width:  500,
		Height: 600,
	})
	view.Run()

	t, err := auth.PollDeviceAuth(resp.DeviceCode)
	if err != nil {
		panic(err)
	}
	if t == nil {
		panic(err)
	}
	return t
}
