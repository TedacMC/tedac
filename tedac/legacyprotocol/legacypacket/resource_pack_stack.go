package legacypacket

import (
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
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
	// BaseGameVersion is the vanilla version that the client should set its resource pack stack to.
	BaseGameVersion string
	// Unsure of the usage.
	Experiments                     []protocol.ExperimentData
	PreviouslyHadExperimentsToggled bool
}

// ID ...
func (*ResourcePackStack) ID() uint32 {
	return packet.IDResourcePackStack
}

// Marshal ...
func (pk *ResourcePackStack) Marshal(w protocol.IO) {
	w.Bool(&pk.TexturePackRequired)
	behaviourLen, textureLen := uint32(len(pk.BehaviourPacks)), uint32(len(pk.TexturePacks))
	w.Varuint32(&behaviourLen)
	for _, pack := range pk.BehaviourPacks {
		legacyprotocol.StackPack(w, &legacyprotocol.StackResourcePack{
			UUID:        pack.UUID,
			Version:     pack.Version,
			SubPackName: pack.SubPackName,
		})
	}
	w.Varuint32(&textureLen)
	for _, pack := range pk.TexturePacks {
		legacyprotocol.StackPack(w, &legacyprotocol.StackResourcePack{
			UUID:        pack.UUID,
			Version:     pack.Version,
			SubPackName: pack.SubPackName,
		})
	}
	w.String(&pk.BaseGameVersion)
	exps := lo.Map(pk.Experiments, func(exp protocol.ExperimentData, _ int) legacyprotocol.ExperimentData {
		return legacyprotocol.ExperimentData{
			Name:    exp.Name,
			Enabled: exp.Enabled,
		}
	})
	legacyprotocol.Experiments(w, &exps)
	w.Bool(&pk.PreviouslyHadExperimentsToggled)
}
