package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

const (
	InventoryTransactionTypeNormal = iota
	InventoryTransactionTypeMismatch
	InventoryTransactionTypeUseItem
	InventoryTransactionTypeUseItemOnEntity
	InventoryTransactionTypeReleaseItem
)

// InventoryTransaction is a packet sent by the client. It essentially exists out of multiple sub-packets,
// each of which have something to do with the inventory in one way or another. Some of these sub-packets
// directly relate to the inventory, others relate to interaction with the world, that could potentially
// result in a change in the inventory.
type InventoryTransaction struct {
	// Actions is a list of actions that took place, that form the inventory transaction together. Each of
	// these actions hold one slot in which one item was changed to another. In general, the combination of
	// all of these actions results in a balanced inventory transaction. This should be checked to ensure that
	// no items are cheated into the inventory.
	Actions []legacyprotocol.InventoryAction
	// TransactionData is a data object that holds data specific to the type of transaction that the
	// TransactionPacket held. Its concrete type must be one of NormalTransactionData, MismatchTransactionData
	// UseItemTransactionData, UseItemOnEntityTransactionData or ReleaseItemTransactionData. If nil is set,
	// the transaction will be assumed to of type InventoryTransactionTypeNormal.
	TransactionData legacyprotocol.InventoryTransactionData
}

// ID ...
func (*InventoryTransaction) ID() uint32 {
	return packet.IDInventoryTransaction
}

// Marshal ...
func (pk *InventoryTransaction) Marshal(w *protocol.Writer) {
	var id uint32
	switch pk.TransactionData.(type) {
	case nil, *legacyprotocol.NormalTransactionData:
		id = InventoryTransactionTypeNormal
	case *legacyprotocol.MismatchTransactionData:
		id = InventoryTransactionTypeMismatch
	case *legacyprotocol.UseItemTransactionData:
		id = InventoryTransactionTypeUseItem
	case *legacyprotocol.UseItemOnEntityTransactionData:
		id = InventoryTransactionTypeUseItemOnEntity
	case *legacyprotocol.ReleaseItemTransactionData:
		id = InventoryTransactionTypeReleaseItem
	}
	w.Varuint32(&id)
	protocol.FuncSlice(w, &pk.Actions, func(action *legacyprotocol.InventoryAction) {
		action.Marshal(w)
	})
	if pk.TransactionData != nil {
		pk.TransactionData.Marshal(w)
	}
}

// Unmarshal ...
func (pk *InventoryTransaction) Unmarshal(r *protocol.Reader) {
	var transactionType uint32
	r.Varuint32(&transactionType)
	protocol.FuncSlice(r, &pk.Actions, func(action *legacyprotocol.InventoryAction) {
		action.Unmarshal(r)
	})
	switch transactionType {
	case InventoryTransactionTypeNormal:
		pk.TransactionData = &legacyprotocol.NormalTransactionData{}
	case InventoryTransactionTypeMismatch:
		pk.TransactionData = &legacyprotocol.MismatchTransactionData{}
	case InventoryTransactionTypeUseItem:
		pk.TransactionData = &legacyprotocol.UseItemTransactionData{}
	case InventoryTransactionTypeUseItemOnEntity:
		pk.TransactionData = &legacyprotocol.UseItemOnEntityTransactionData{}
	case InventoryTransactionTypeReleaseItem:
		pk.TransactionData = &legacyprotocol.ReleaseItemTransactionData{}
	default:
		r.UnknownEnumOption(transactionType, "inventory transaction type")
	}
	pk.TransactionData.Unmarshal(r)
}
