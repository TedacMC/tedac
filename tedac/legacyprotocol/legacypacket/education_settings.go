package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// EducationSettings is a packet sent by the server to update Minecraft: Education Edition related settings.
// It is unused by the normal base game.
type EducationSettings struct {
	// CodeBuilderDefaultURI is the default URI that the code builder is ran on. Using this, a Code Builder
	// program can make code directly affect the server.
	CodeBuilderDefaultURI string
	// CodeBuilderTitle is the title of the code builder shown when connected to the CodeBuilderDefaultURI.
	CodeBuilderTitle string
	// CanResizeCodeBuilder specifies if clients connected to the world should be able to resize the code
	// builder when it is opened.
	CanResizeCodeBuilder bool
	// OverrideURI ...
	OverrideURI protocol.Optional[string]
	// HasQuiz specifies if the world has a quiz connected to it.
	HasQuiz bool
}

// ID ...
func (*EducationSettings) ID() uint32 {
	return packet.IDEducationSettings
}

// Marshal ...
func (pk *EducationSettings) Marshal(w protocol.IO) {
	w.String(&pk.CodeBuilderDefaultURI)
	w.String(&pk.CodeBuilderTitle)
	w.Bool(&pk.CanResizeCodeBuilder)
	protocol.OptionalFunc(w, &pk.OverrideURI, w.String)
	w.Bool(&pk.HasQuiz)
}
