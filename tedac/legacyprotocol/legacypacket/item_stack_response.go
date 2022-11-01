package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

// ItemStackResponse is sent by the server in response to an ItemStackRequest packet from the client. This
// packet is used to either approve or reject ItemStackRequests from the client. If a request is approved, the
// client will simply continue as normal. If rejected, the client will undo the actions so that the inventory
// should be in sync with the server again.
type ItemStackResponse struct {
	// Responses is a list of responses to ItemStackRequests sent by the client before. Responses either
	// approve or reject a request from the client.
	// Vanilla limits the size of this slice to 4096.
	Responses []legacyprotocol.ItemStackResponse
}

// ID ...
func (*ItemStackResponse) ID() uint32 {
	return packet.IDItemStackResponse
}

// Marshal ...
func (pk *ItemStackResponse) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Responses))
	w.Varuint32(&l)
	for _, resp := range pk.Responses {
		legacyprotocol.WriteStackResponse(w, &resp)
	}
}

// Unmarshal ...
func (pk *ItemStackResponse) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	pk.Responses = make([]legacyprotocol.ItemStackResponse, count)
	for i := uint32(0); i < count; i++ {
		legacyprotocol.StackResponse(r, &pk.Responses[i])
	}
}
