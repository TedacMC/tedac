package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	PlayerActionStartBreak = iota
	PlayerActionAbortBreak
	PlayerActionStopBreak
	PlayerActionGetUpdatedBlock
	PlayerActionDropItem
	PlayerActionStartSleeping
	PlayerActionStopSleeping
	PlayerActionRespawn
	PlayerActionJump
	PlayerActionStartSprint
	PlayerActionStopSprint
	PlayerActionStartSneak
	PlayerActionStopSneak
	PlayerActionCreativePlayerDestroyBlock
	PlayerActionDimensionChangeDone
	PlayerActionStartGlide
	PlayerActionStopGlide
	PlayerActionBuildDenied
	PlayerActionCrackBreak
	PlayerActionChangeSkin
	PlayerActionSetEnchantmentSeed
	PlayerActionStartSwimming
	PlayerActionStopSwimming
	PlayerActionStartSpinAttack
	PlayerActionStopSpinAttack
	PlayerActionStartBuildingBlock
)

// PlayerAction is sent by the client when it executes any action, for example starting to sprint, swim,
// starting the breaking of a block, dropping an item, etc.
type PlayerAction struct {
	// EntityNetworkID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// ActionType is the ID of the action that was executed by the player. It is one of the constants that may
	// be found above.
	ActionType int32
	// BlockPosition is the position of the target block, if the action with the ActionType set concerned a
	// block. If that is not the case, the block position will be zero.
	BlockPosition protocol.BlockPos
	// BlockFace is the face of the target block that was touched. If the action with the ActionType set
	// concerned a block. If not, the face is always 0.
	BlockFace int32
}

// ID ...
func (*PlayerAction) ID() uint32 {
	return packet.IDPlayerAction
}

// Marshal ...
func (pk *PlayerAction) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.ActionType)
	w.UBlockPos(&pk.BlockPosition)
	w.Varint32(&pk.BlockFace)
}

// Unmarshal ...
func (pk *PlayerAction) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Varint32(&pk.ActionType)
	r.UBlockPos(&pk.BlockPosition)
	r.Varint32(&pk.BlockFace)
}
