package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ResourcePackStack is sent by the server to send the order in which resource packs and behaviour packs
// should be applied (and downloaded) by the client.
type ResourcePackStack struct {
	// TexturePackRequired specifies if the client must accept the texture packs the server has in order to
	// join the server. If set to true, the client gets the option to either download the resource packs and
	// join, or quit entirely. Behaviour packs never have to be downloaded.
	TexturePackRequired bool
	// BehaviourPack is a list of behaviour packs that the client needs to download before joining the server.
	// All of these behaviour packs will be applied together, and the order does not necessarily matter.
	BehaviourPacks []protocol.StackResourcePack
	// TexturePacks is a list of texture packs that the client needs to download before joining the server.
	// The order of these texture packs specifies the order that they are applied in on the client side. The
	// first in the list will be applied first.
	TexturePacks []protocol.StackResourcePack
	// Experimental specifies if the resource packs in the stack are experimental. This is internal and should
	// always be set to false.
	Experimental bool
}

// ID ...
func (*ResourcePackStack) ID() uint32 {
	return packet.IDResourcePackStack
}

// Marshal ...
func (pk *ResourcePackStack) Marshal(w *protocol.Writer) {
	w.Bool(&pk.TexturePackRequired)
	protocol.Slice(w, &pk.BehaviourPacks)
	protocol.Slice(w, &pk.TexturePacks)
	w.Bool(&pk.Experimental)
}

// Unmarshal ...
func (pk *ResourcePackStack) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.TexturePackRequired)
	protocol.Slice(r, &pk.BehaviourPacks)
	protocol.Slice(r, &pk.TexturePacks)
	r.Bool(&pk.Experimental)
}
