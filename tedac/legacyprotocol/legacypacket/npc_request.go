package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// NPCRequest is sent by the client when it interacts with an NPC.
// The packet is specifically made for Education Edition, where NPCs are available to use.
type NPCRequest struct {
	// EntityRuntimeID is the runtime ID of the NPC entity that the player interacted with. It is the same
	// as sent by the server when spawning the entity.
	EntityRuntimeID uint64
	// RequestType is the type of the request, which depends on the permission that the player has. It will
	// be either a type that indicates that the NPC should show its dialog, or that it should open the
	// editing window.
	RequestType byte
	// CommandString is the command string set in the NPC. It may consist of multiple commands, depending on
	// what the player set in it.
	CommandString string
	// ActionType is the type of the action to execute.
	ActionType byte
}

// ID ...
func (*NPCRequest) ID() uint32 {
	return packet.IDNPCRequest
}

// Marshal ...
func (pk *NPCRequest) Marshal(w protocol.IO) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Uint8(&pk.RequestType)
	w.String(&pk.CommandString)
	w.Uint8(&pk.ActionType)
}
