package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NetworkSettings is sent by the server to update a variety of network settings. These settings modify the
// way packets are sent over the network stack.
type NetworkSettings struct {
	// CompressionThreshold is the minimum size of a packet that is compressed when sent. If the size of a
	// packet is under this value, it is not compressed.
	// When set to 0, all packets will be left uncompressed.
	CompressionThreshold uint16
}

// ID ...
func (*NetworkSettings) ID() uint32 {
	return packet.IDNetworkSettings
}

// Marshal ...
func (pk *NetworkSettings) Marshal(w protocol.IO) {
	w.Uint16(&pk.CompressionThreshold)
}
