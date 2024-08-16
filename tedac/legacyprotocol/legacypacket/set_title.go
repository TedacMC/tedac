package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SetTitle is sent by the server to make a title, subtitle or action bar shown to a player. It has several
// fields that allow setting the duration of the titles.
type SetTitle struct {
	// ActionType is the type of the action that should be executed upon the title of a player. It is one of
	// the constants above and specifies the response of the client to the packet.
	ActionType int32
	// Text is the text of the title, which has a different meaning depending on the ActionType that the
	// packet has. The text is the text of a title, subtitle or action bar, depending on the type set.
	Text string
	// FadeInDuration is the duration that the title takes to fade in on the screen of the player. It is
	// measured in 20ths of a second (AKA in ticks).
	FadeInDuration int32
	// RemainDuration is the duration that the title remains on the screen of the player. It is measured in
	// 20ths of a second (AKA in ticks).
	RemainDuration int32
	// FadeOutDuration is the duration that the title takes to fade out of the screen of the player. It is
	// measured in 20ths of a second (AKA in ticks).
	FadeOutDuration int32
}

// ID ...
func (*SetTitle) ID() uint32 {
	return packet.IDSetTitle
}

func (pk *SetTitle) Marshal(io protocol.IO) {
	io.Varint32(&pk.ActionType)
	io.String(&pk.Text)
	io.Varint32(&pk.FadeInDuration)
	io.Varint32(&pk.RemainDuration)
	io.Varint32(&pk.FadeOutDuration)
}
