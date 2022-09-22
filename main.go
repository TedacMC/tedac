package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac"
	"log"
	"os"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	conf := readConfig()
	src := tokenSource()

	p, err := minecraft.NewForeignStatusProvider(conf.Connection.RemoteAddress)
	if err != nil {
		panic(err)
	}
	listener, err := minecraft.ListenConfig{
		StatusProvider:    p,
		AcceptedProtocols: []minecraft.Protocol{tedac.Protocol{}},
	}.Listen("raknet", conf.Connection.LocalAddress)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Tedac is now running on " + conf.Connection.LocalAddress + " for v" + protocol.CurrentVersion)
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(c.(*minecraft.Conn), listener, &conf, src)
	}
}

// defaultSkinResourcePatch holds the skin resource patch assigned to a player when they wear a custom skin.
const defaultSkinResourcePatch = `{
   "geometry" : {
      "default" : "geometry.humanoid.custom"
   }
}
`

// handleConn handles a new incoming minecraft.Conn from the minecraft.Listener passed.
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, config *config, src oauth2.TokenSource) {
	clientData := conn.ClientData()
	if _, ok := conn.Protocol().(tedac.Protocol); ok { // TODO: Adjust this inside Protocol itself.
		clientData.GameVersion = protocol.CurrentVersion

		clientData.SkinResourcePatch = base64.StdEncoding.EncodeToString([]byte(defaultSkinResourcePatch))
		clientData.SkinImageHeight = 64
		clientData.SkinImageWidth = 64
	}

	serverConn, err := minecraft.Dialer{
		TokenSource: src,
		ClientData:  clientData,
	}.Dial("raknet", config.Connection.RemoteAddress)
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
		defer listener.Disconnect(conn, "connection lost")
		defer serverConn.Close()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				return
			}
			if err := serverConn.WritePacket(pk); err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = listener.Disconnect(conn, disconnect.Error())
				}
				return
			}
		}
	}()
	go func() {
		defer serverConn.Close()
		defer listener.Disconnect(conn, "connection lost")
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				if disconnect, ok := errors.Unwrap(err).(minecraft.DisconnectError); ok {
					_ = listener.Disconnect(conn, disconnect.Error())
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
				config.Connection.RemoteAddress = fmt.Sprintf("%s:%d", pk.Address, pk.Port)

				pk.Address = "127.0.0.1"
				pk.Port = 19132

				fmt.Println(config.Connection.RemoteAddress)
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
		}
	}()
}

type config struct {
	Connection struct {
		LocalAddress  string
		RemoteAddress string
	}
}

func readConfig() config {
	c := config{}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, err := os.Create("config.toml")
		if err != nil {
			log.Fatalf("error creating config: %v", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("error encoding default config: %v", err)
		}
		if _, err := f.Write(data); err != nil {
			log.Fatalf("error writing encoded default config: %v", err)
		}
		_ = f.Close()
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	if c.Connection.LocalAddress == "" {
		c.Connection.LocalAddress = "0.0.0.0:19132"
	}
	data, _ = toml.Marshal(c)
	if err := os.WriteFile("config.toml", data, 0644); err != nil {
		log.Fatalf("error writing config file: %v", err)
	}
	return c
}

// tokenSource returns a token source for using with a gophertunnel client. It either reads it from the
// token.tok file if cached or requests logging in with a device code.
func tokenSource() oauth2.TokenSource {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	token := new(oauth2.Token)
	tokenData, err := os.ReadFile("token.tok")
	if err == nil {
		_ = json.Unmarshal(tokenData, token)
	} else {
		token, err = auth.RequestLiveToken()
		check(err)
	}
	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = auth.RequestLiveToken()
		check(err)
		src = auth.RefreshTokenSource(token)
	}
	tok, _ := src.Token()
	b, _ := json.Marshal(tok)
	_ = os.WriteFile("token.tok", b, 0644)
	return src
}
