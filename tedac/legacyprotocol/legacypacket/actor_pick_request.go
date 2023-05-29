package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ActorPickRequest is sent by the client when it tries to pick an entity, so that it gets a spawn egg which
// can spawn that entity.
type ActorPickRequest struct {
	// EntityUniqueID is the unique ID of the entity that was attempted to be picked. The server must find the
	// type of that entity and provide the correct spawn egg to the player.
	EntityUniqueID int64
	// HotBarSlot is the held hot bar slot of the player at the time of trying to pick the entity. If empty,
	// the resulting spawn egg should be put into this slot.
	HotBarSlot byte
}

// ID ...
func (*ActorPickRequest) ID() uint32 {
	return packet.IDActorPickRequest
}

// Marshal ...
func (pk *ActorPickRequest) Marshal(w protocol.IO) {
	w.Int64(&pk.EntityUniqueID)
	w.Uint8(&pk.HotBarSlot)
}
