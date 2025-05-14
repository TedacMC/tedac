package legacypacket

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// EntityFall is sent by the client when it falls from a distance onto a block that would damage the player.
type EntityFall struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// FallDistance is the distance that the entity fell until it hit the ground. The damage would otherwise
	// be calculated using this field.
	FallDistance float32
	// InVoid specifies if the fall was in the void. The player can't fall below roughly Y=-40.
	InVoid bool
}

const IDEntityFall = 37

// ID ...
func (*EntityFall) ID() uint32 {
	return IDEntityFall
}

func (pk *EntityFall) Marshal(io protocol.IO) {
	io.Uint64(&pk.EntityRuntimeID)
	io.Float32(&pk.FallDistance)
	io.Bool(&pk.InVoid)
}
