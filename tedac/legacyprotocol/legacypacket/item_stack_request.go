package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

// ItemStackRequest is sent by the client to change item stacks in an inventory. It is essentially a
// replacement of the InventoryTransaction packet added in 1.16 for inventory specific actions, such as moving
// items around or crafting. The InventoryTransaction packet is still used for actions such as placing blocks
// and interacting with entities.
type ItemStackRequest struct {
	// Requests holds a list of item stack requests. These requests are all separate, but the client buffers
	// the requests, so you might find multiple unrelated requests in this packet.
	Requests []legacyprotocol.ItemStackRequest
}

// ID ...
func (*ItemStackRequest) ID() uint32 {
	return packet.IDItemStackRequest
}

// Marshal ...
func (pk *ItemStackRequest) Marshal(w protocol.IO) {
	l := uint32(len(pk.Requests))
	w.Varuint32(&l)
	for _, req := range pk.Requests {
		legacyprotocol.WriteStackRequest(w, &req)
	}
}
