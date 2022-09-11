package tedac

import (
	"fmt"
	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"net"
)

// MultiRakNet is an implementation of a RakNet v9/10 Network.
type MultiRakNet struct {
	minecraft.RakNet
}

// legacyRakNet represents the legacy version of RakNet, necessary for v1.12.0.
const legacyRakNet = 9

// Listen ...
func (MultiRakNet) Listen(address string) (minecraft.NetworkListener, error) {
	return raknet.ListenConfig{
		ProtocolVersions: []byte{legacyRakNet}, // Version 9 is required for v1.12.0 MV.
	}.Listen(address)
}

// Compression ...
func (MultiRakNet) Compression(conn net.Conn) packet.Compression {
	fmt.Println(conn.(*raknet.Conn).ProtocolVersion())
	switch conn.(*raknet.Conn).ProtocolVersion() {
	case legacyRakNet:
		return zLibCompression{}
	}
	return packet.FlateCompression{}
}

// init registers the MultiRakNet network. It overrides the existing minecraft.RakNet network.
func init() {
	minecraft.RegisterNetwork("raknet", MultiRakNet{})
}
