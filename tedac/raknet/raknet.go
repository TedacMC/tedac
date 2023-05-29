package raknet

import (
	"net"

	"github.com/sandertv/go-raknet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// MultiRakNet is an implementation of a RakNet v9/10 Network.
type MultiRakNet struct {
	minecraft.RakNet
}

// legacyRakNet represents the legacy version of RakNet, necessary for v1.16.100.
const legacyRakNet = 10

// Listen ...
func (MultiRakNet) Listen(address string) (minecraft.NetworkListener, error) {
	return raknet.ListenConfig{
		ProtocolVersions: []byte{10}, // Version 9 is required for v1.16.100 MV.
	}.Listen(address)
}

// Compression ...
func (MultiRakNet) Compression(conn net.Conn) packet.Compression {
	if conn.(*raknet.Conn).ProtocolVersion() == legacyRakNet {
		return packet.FlateCompression
	}
	return packet.SnappyCompression
}

// init registers the MultiRakNet network. It overrides the existing minecraft.RakNet network.
func init() {
	minecraft.RegisterNetwork("raknet", MultiRakNet{})
}
