package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// The following program implements a proxy that forwards players from one local address to a remote address.

const (
	GUI        = false
	RPC        = true
	DISCORD_ID = "710885082100924416"
)

func main() {
	if GUI {
		run()
	} else {
		server(readConfig().Connection.LocalAddress, readConfig().Connection.RemoteAddress)
	}
}

func server(local string, remote string) error {
	conf := readConfig()
	src := tokenSource()

	if RPC {
		go rpc()
	}

	p, err := minecraft.NewForeignStatusProvider(conf.Connection.RemoteAddress)
	if err != nil {
		return err
	}
	listener, err := minecraft.ListenConfig{
		StatusProvider:    p,
		AcceptedProtocols: []minecraft.Protocol{tedac.Protocol{}},
	}.Listen("raknet", local)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Tedac is now running on %s\n", local)
	for {
		c, err := listener.Accept()
		if err != nil {
			return err
		}
		go handleConn(c.(*minecraft.Conn), listener, &conf, src, remote)
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
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, config *config, src oauth2.TokenSource, remote string) {
	clientData := conn.ClientData()
	if _, ok := conn.Protocol().(tedac.Protocol); ok {
		clientData.GameVersion = protocol.CurrentVersion

		clientData.SkinResourcePatch = base64.StdEncoding.EncodeToString([]byte(defaultSkinResourcePatch))
		clientData.SkinImageHeight = 64
		clientData.SkinImageWidth = 64
	}

	serverConn, err := minecraft.Dialer{
		TokenSource: src,
		ClientData:  clientData,
	}.Dial("raknet", remote)
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
			case *packet.Transfer:
				address := strings.Split(config.Connection.LocalAddress, ":")
				port, _ := strconv.Atoi(address[1])

				pk.Address = address[0]
				pk.Port = uint16(port)
				config.Connection.RemoteAddress = fmt.Sprintf("%s:%d", pk.Address, pk.Port)
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
	tokenData, err := ioutil.ReadFile("token.tok")
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
	_ = ioutil.WriteFile("token.tok", b, 0644)
	return src
}
