package tedac_test

import (
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/tedacmc/tedac"
	"testing"
)

func TestListen(t *testing.T) {
	// Listen on the address with port 19132.
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

			// Make the client spawn in the world using conn.StartGame. An error is returned if the client
			// times out during the connection.
			if err := conn.StartGame(minecraft.GameData{}); err != nil {
				return
			}

			for {
				// Read a packet from the connection: ReadPacket returns an error if the connection is closed or if
				// a read timeout is set. You will generally want to return or break if this happens.
				_, err := conn.ReadPacket()
				if err != nil {
					break
				}
			}
		}()
	}
}
