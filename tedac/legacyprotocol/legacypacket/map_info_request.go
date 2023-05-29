package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// MapInfoRequest is sent by the client to request the server to deliver information of a certain map in the
// inventory of the player. The server should respond with a ClientBoundMapItemData packet.
type MapInfoRequest struct {
	// MapID is the unique identifier that represents the map that is requested over network. It remains
	// consistent across sessions.
	MapID int64
}

// ID ...
func (*MapInfoRequest) ID() uint32 {
	return packet.IDMapInfoRequest
}

// Marshal ...
func (pk *MapInfoRequest) Marshal(w protocol.IO) {
	w.Varint64(&pk.MapID)
}
