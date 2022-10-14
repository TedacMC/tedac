package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

// InventoryContent is sent by the server to update the full content of a particular inventory. It is usually
// sent for the main inventory of the player, but also works for other inventories that are currently opened
// by the player.
type InventoryContent struct {
	// WindowID is the ID that identifies one of the windows that the client currently has opened, or one of
	// the consistent windows such as the main inventory.
	WindowID uint32
	// Content is the new content of the inventory. The length of this slice must be equal to the full size of
	// the inventory window updated.
	Content []legacyprotocol.ItemStack
}

// ID ...
func (*InventoryContent) ID() uint32 {
	return packet.IDInventoryContent
}

// Marshal ...
func (pk *InventoryContent) Marshal(w *protocol.Writer) {
	w.Varuint32(&pk.WindowID)
	protocol.FuncSlice(w, &pk.Content, func(stack *legacyprotocol.ItemStack) {
		legacyprotocol.WriteItem(w, stack)
	})
}

// Unmarshal ...
func (pk *InventoryContent) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&pk.WindowID)
	protocol.FuncSlice(r, &pk.Content, func(stack *legacyprotocol.ItemStack) {
		legacyprotocol.Item(r, stack)
	})
}
