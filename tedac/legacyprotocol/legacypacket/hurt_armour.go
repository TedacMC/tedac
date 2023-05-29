package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// HurtArmour is sent by the server to damage the player's armour after being hit. The packet should never be
// used by servers as it hands the responsibility over to the player completely, while the server can easily
// reliably update the armour damage of players itself.
type HurtArmour struct {
	// Cause is the cause of the damage dealt to the armour.
	Cause int32
	// Damage is the amount of damage points that was dealt to the player. The damage to the armour will be
	// calculated by the client based upon this damage, and will also be based upon any enchantments like
	// thorns that the armour may have.
	Damage int32
}

// ID ...
func (*HurtArmour) ID() uint32 {
	return packet.IDHurtArmour
}

// Marshal ...
func (pk *HurtArmour) Marshal(w protocol.IO) {
	w.Varint32(&pk.Cause)
	w.Varint32(&pk.Damage)
}
