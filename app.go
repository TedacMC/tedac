package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac"
	"github.com/wailsapp/wails/lib/renderer/webview"
	"golang.org/x/oauth2"
	"net"
	"os"
	"sync"
)

// App ...
type App struct {
	listener      *minecraft.Listener
	remoteAddress string
	localPort     uint16

	src oauth2.TokenSource
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{src: tokenSource()}
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

	a.remoteAddress = address
	a.localPort = uint16(port)

	a.listener, err = minecraft.ListenConfig{
		StatusProvider:    p,
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

// startup is called when the app starts. The context is saved, so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// defaultSkinResourcePatch holds the skin resource patch assigned to a player when they wear a custom skin.
const defaultSkinResourcePatch = `{
   "geometry" : {
      "default" : "geometry.humanoid.custom"
   }
}
`

// handleConn handles a new incoming minecraft.Conn from the minecraft.Listener passed.
func (a *App) handleConn(conn *minecraft.Conn) {
	clientData := conn.ClientData()
	if _, ok := conn.Protocol().(tedac.Protocol); ok { // TODO: Adjust this inside Protocol itself.
		clientData.GameVersion = protocol.CurrentVersion

		clientData.SkinResourcePatch = base64.StdEncoding.EncodeToString([]byte(defaultSkinResourcePatch))
		clientData.SkinImageHeight = 64
		clientData.SkinImageWidth = 64
	}

	serverConn, err := minecraft.Dialer{
		TokenSource: a.src,
		ClientData:  clientData,
	}.Dial("raknet", a.remoteAddress)
	if err != nil {
		panic(err)
	}
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
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

	go func() {
		defer a.listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			if err := serverConn.WritePacket(pk); err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
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
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = a.listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
			switch pk := pk.(type) {
			case *packet.SubChunk:
				// TODO: Re-encode the sub chunk to the correct format.
			case *packet.LevelChunk:
				if pk.SubChunkRequestMode != protocol.SubChunkRequestModeLegacy {
					max := world.Overworld.Range().Height() >> 4
					if pk.SubChunkRequestMode == protocol.SubChunkRequestModeLimited {
						max = int(pk.HighestSubChunk)
					}

					offsets := make([]protocol.SubChunkOffset, 0, max)
					for i := 0; i < max; i++ {
						offsets = append(offsets, protocol.SubChunkOffset{0, int8(i), 0})
					}

					_ = serverConn.WritePacket(&packet.SubChunkRequest{
						Position: protocol.SubChunkPos{pk.Position.X(), 0, pk.Position.Z()},
						Offsets:  offsets,
					})
					continue
				}
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
		token, err = requestToken()
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = requestToken()
		if err != nil {
			panic(err)
		}
		src = auth.RefreshTokenSource(token)
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile("token.tok", b, 0644)
	return src
}

// requestToken opens a new WebView2 window and requests the user to log in. The token is returned if successful.
func requestToken() (*oauth2.Token, error) {
	resp, err := auth.StartDeviceAuth()
	if err != nil {
		return nil, err
	}
	webview.NewWebview(webview.Settings{
		Title:  "Tedac Authentication",
		URL:    "https://login.live.com/oauth20_remoteconnect.srf?lc=1033&otc=" + resp.UserCode,
		Width:  500,
		Height: 600,
	}).Run()

	t, err := auth.PollDeviceAuth(resp.DeviceCode)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("no token received")
	}
	return t, nil
}
