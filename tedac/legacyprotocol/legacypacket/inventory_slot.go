package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

// InventorySlot is sent by the server to update a single slot in one of the inventory windows that the client
// currently has opened. Usually this is the main inventory, but it may also be the off hand or, for example,
// a chest inventory.
type InventorySlot struct {
	// WindowID is the ID of the window that the packet modifies. It must point to one of the windows that the
	// client currently has opened.
	WindowID uint32
	// Slot is the index of the slot that the packet modifies. The new item will be set to the slot at this
	// index.
	Slot uint32
	// NewItem is the item to be put in the slot at Slot. It will overwrite any item that may currently
	// be present in that slot.
	NewItem legacyprotocol.ItemInstance
}

// ID ...
func (*InventorySlot) ID() uint32 {
	return packet.IDInventorySlot
}

// Marshal ...
func (pk *InventorySlot) Marshal(w protocol.IO) {
	w.Varuint32(&pk.WindowID)
	w.Varuint32(&pk.Slot)
	legacyprotocol.WriteItemInst(w, &pk.NewItem)
}
