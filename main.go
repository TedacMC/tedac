package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/tedacmc/tedac/tedac"
	"golang.org/x/oauth2"
)

func main() {
	tok := tokenSource()
	listener, err := minecraft.ListenConfig{
		StatusProvider:    minecraft.NewStatusProvider("Tedac Listen Test"),
		AcceptedProtocols: []minecraft.Protocol{tedac.Protocol{}},
	}.Listen("raknet", ":19132")
	if err != nil {
		panic(err)
	}

	for {
		// Accept connections in a for loop. Accept will only return an error if the minecraft.Listener is
		// closed. (So never unexpectedly.)
		c, err := listener.Accept()
		if err != nil {
			return
		}
		conn := c.(*minecraft.Conn)

		go func() {
			// Process the connection on another goroutine as you would with TCP connections.
			defer conn.Close()
			clientData := conn.ClientData()
			serverConn, err := minecraft.Dialer{
				TokenSource: tok,
				ClientData:  clientData,
			}.Dial("raknet", "127.0.0.1:19133")

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
					if err := conn.WritePacket(pk); err != nil {
						return
					}
				}
			}()
		}()
	}
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
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	if c.Connection.LocalAddress == "" {
		c.Connection.LocalAddress = "0.0.0.0:19132"
	}
	if c.Connection.RemoteAddress == "" {
		c.Connection.RemoteAddress = "vasar.land:19132"
	}
	data, _ = toml.Marshal(c)
	if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
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
