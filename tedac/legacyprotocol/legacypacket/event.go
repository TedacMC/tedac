package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	EventAchievementAwarded = iota
	EventEntityInteract
	EventPortalBuilt
	EventPortalUsed
	EventMobKilled
	EventCauldronUsed
	EventPlayerDeath
	EventBossKilled
	EventAgentCommand
	EventAgentCreated
	EventBannerPatternRemoved
	EventCommandExecuted
	EventFishBucketed
)

// Event is sent by the server to send an event with additional data. It is typically sent to the client for
// telemetry reasons, much like the SimpleEvent packet.
type Event struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EventType is the type of the event to be called. It is one of the constants that may be found above.
	EventType int32
	// UsePlayerID ... TODO: Figure out what this is for.
	UsePlayerID byte
}

// ID ...
func (*Event) ID() uint32 {
	return packet.IDEvent
}

// Marshal ...
func (pk *Event) Marshal(w protocol.IO) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.EventType)
	w.Uint8(&pk.UsePlayerID)

	// TODO: Add fields for all Event types.
}
