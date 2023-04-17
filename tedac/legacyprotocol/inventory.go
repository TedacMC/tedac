package legacyprotocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	InventoryActionSourceContainer = 0
	InventoryActionSourceWorld     = 2
	InventoryActionSourceCreative  = 3
	InventoryActionSourceTODO      = 99999
)

const (
	WindowIDInventory = 0
	WindowIDOffHand   = 119
	WindowIDArmour    = 120
	WindowIDUI        = 124
)

// InventoryAction represents a single action that took place during an inventory transaction. On itself, this
// inventory action is always unbalanced: It must be combined with other actions in an inventory transaction
// to form a balanced transaction.
type InventoryAction struct {
	// SourceType is the source type of the inventory action. It is one of the constants above.
	SourceType uint32
	// WindowID is the ID of the window that the client has opened. The window ID is not set if the SourceType
	// is InventoryActionSourceWorld.
	WindowID int32
	// SourceFlags is a combination of flags that is only set if the SourceType is InventoryActionSourceWorld.
	SourceFlags uint32
	// InventorySlot is the slot in which the action took place. Each action only describes the change of item
	// in a single slot.
	InventorySlot uint32
	// OldItem is the item that was present in the slot before the inventory action. It should be checked by
	// the server to ensure the inventories were not out of sync.
	OldItem ItemStack
	// NewItem is the new item that was put in the InventorySlot that the OldItem was in. It must be checked
	// in combination with other inventory actions to ensure that the transaction is balanced.
	NewItem ItemStack
}

// Marshal encodes/decodes an InventoryAction.
func (x *InventoryAction) Marshal(io protocol.IO) {
	io.Varuint32(&x.SourceType)
	switch x.SourceType {
	case InventoryActionSourceContainer, InventoryActionSourceTODO:
		io.Varint32(&x.WindowID)
	case InventoryActionSourceWorld:
		io.Varuint32(&x.SourceFlags)
	}
	io.Varuint32(&x.InventorySlot)
	Item(io, &x.OldItem)
	Item(io, &x.NewItem)
}

// InventoryTransactionData represents an object that holds data specific to an inventory transaction type.
// The data it holds depends on the type.
type InventoryTransactionData interface {
	// Marshal encodes a serialised inventory transaction data object.
	Marshal(w *protocol.Writer)
	// Unmarshal decodes a serialised inventory transaction data object.
	Unmarshal(r *protocol.Reader)
}

// NormalTransactionData represents an inventory transaction data object for normal transactions, such as
// crafting. It has no content.
type NormalTransactionData struct{}

// MismatchTransactionData represents a mismatched inventory transaction's data object.
type MismatchTransactionData struct{}

const (
	UseItemActionClickBlock = iota
	UseItemActionClickAir
	UseItemActionBreakBlock
)

// UseItemTransactionData represents an inventory transaction data object sent when the client uses an item on
// a block.
type UseItemTransactionData struct {
	// ActionType is the type of the UseItem inventory transaction. It is one of the action types found above,
	// and specifies the way the player interacted with the block.
	ActionType uint32
	// BlockPosition is the position of the block that was interacted with. This is only really a correct
	// block position if ActionType is not UseItemActionClickAir.
	BlockPosition protocol.BlockPos
	// BlockFace is the face of the block that was interacted with. When clicking the block, it is the face
	// clicked. When breaking the block, it is the face that was last being hit until the block broke.
	BlockFace int32
	// HotBarSlot is the hot bar slot that the player was holding while clicking the block. It should be used
	// to ensure that the hot bar slot and held item are correctly synchronised with the server.
	HotBarSlot int32
	// HeldItem is the item that was held to interact with the block. The server should check if this item
	// is actually present in the HotBarSlot.
	HeldItem ItemStack
	// Position is the position of the player at the time of interaction. For clicking a block, this is the
	// position at that time, whereas for breaking the block it is the position at the time of breaking.
	Position mgl32.Vec3
	// ClickedPosition is the position that was clicked relative to the block's base coordinate. It can be
	// used to find out exactly where a player clicked the block.
	ClickedPosition mgl32.Vec3
	// BlockRuntimeID is the runtime ID of the block that was clicked. It may be used by the server to verify
	// that the player's world client-side is synchronised with the server's.
	BlockRuntimeID uint32
}

const (
	UseItemOnEntityActionInteract = iota
	UseItemOnEntityActionAttack
)

// UseItemOnEntityTransactionData represents an inventory transaction data object sent when the client uses
// an item on an entity.
type UseItemOnEntityTransactionData struct {
	// TargetEntityRuntimeID is the entity runtime ID of the target that was clicked. It is the runtime ID
	// that was assigned to it in the AddEntity packet.
	TargetEntityRuntimeID uint64
	// ActionType is the type of the UseItemOnEntity inventory transaction. It is one of the action types
	// found in the constants above, and specifies the way the player interacted with the entity.
	ActionType uint32
	// HotBarSlot is the hot bar slot that the player was holding while clicking the entity. It should be used
	// to ensure that the hot bar slot and held item are correctly synchronised with the server.
	HotBarSlot int32
	// HeldItem is the item that was held to interact with the entity. The server should check if this item
	// is actually present in the HotBarSlot.
	HeldItem ItemStack
	// Position is the position of the player at the time of clicking the entity.
	Position mgl32.Vec3
	// ClickedPosition is the position that was clicked relative to the entity's base coordinate. It can be
	// used to find out exactly where a player clicked the entity.
	ClickedPosition mgl32.Vec3
}

const (
	ReleaseItemActionRelease = iota
	ReleaseItemActionConsume
)

// ReleaseItemTransactionData represents an inventory transaction data object sent when the client releases
// the item it was using, for example when stopping while eating or stopping the charging of a bow.
type ReleaseItemTransactionData struct {
	// ActionType is the type of the ReleaseItem inventory transaction. It is one of the action types found
	// in the constants above, and specifies the way the item was released.
	ActionType uint32
	// HotBarSlot is the hot bar slot that the player was holding while releasing the item. It should be used
	// to ensure that the hot bar slot and held item are correctly synchronised with the server.
	HotBarSlot int32
	// HeldItem is the item that was released. The server should check if this item is actually present in the
	// HotBarSlot.
	HeldItem ItemStack
	// HeadPosition is the position of the player's head at the time of releasing the item. This is used
	// mainly for purposes such as spawning eating particles at that position.
	HeadPosition mgl32.Vec3
}

// Marshal ...
func (data *UseItemTransactionData) Marshal(w *protocol.Writer) {
	w.Varuint32(&data.ActionType)
	w.UBlockPos(&data.BlockPosition)
	w.Varint32(&data.BlockFace)
	w.Varint32(&data.HotBarSlot)
	WriteItem(w, &data.HeldItem)
	w.Vec3(&data.Position)
	w.Vec3(&data.ClickedPosition)
	w.Varuint32(&data.BlockRuntimeID)
}

// Unmarshal ...
func (data *UseItemTransactionData) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&data.ActionType)
	r.UBlockPos(&data.BlockPosition)
	r.Varint32(&data.BlockFace)
	r.Varint32(&data.HotBarSlot)
	Item(r, &data.HeldItem)
	r.Vec3(&data.Position)
	r.Vec3(&data.ClickedPosition)
	r.Varuint32(&data.BlockRuntimeID)
}

// Marshal ...
func (data *UseItemOnEntityTransactionData) Marshal(w *protocol.Writer) {
	w.Varuint64(&data.TargetEntityRuntimeID)
	w.Varuint32(&data.ActionType)
	w.Varint32(&data.HotBarSlot)
	WriteItem(w, &data.HeldItem)
	w.Vec3(&data.Position)
	w.Vec3(&data.ClickedPosition)
}

// Unmarshal ...
func (data *UseItemOnEntityTransactionData) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&data.TargetEntityRuntimeID)
	r.Varuint32(&data.ActionType)
	r.Varint32(&data.HotBarSlot)
	Item(r, &data.HeldItem)
	r.Vec3(&data.Position)
	r.Vec3(&data.ClickedPosition)
}

// Marshal ...
func (data *ReleaseItemTransactionData) Marshal(w *protocol.Writer) {
	w.Varuint32(&data.ActionType)
	w.Varint32(&data.HotBarSlot)
	WriteItem(w, &data.HeldItem)
	w.Vec3(&data.HeadPosition)
}

// Unmarshal ...
func (data *ReleaseItemTransactionData) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&data.ActionType)
	r.Varint32(&data.HotBarSlot)
	Item(r, &data.HeldItem)
	r.Vec3(&data.HeadPosition)
}

// Marshal ...
func (*NormalTransactionData) Marshal(*protocol.Writer) {}

// Unmarshal ...
func (*NormalTransactionData) Unmarshal(*protocol.Reader) {}

// Marshal ...
func (*MismatchTransactionData) Marshal(*protocol.Writer) {}

// Unmarshal ...
func (*MismatchTransactionData) Unmarshal(*protocol.Reader) {}
