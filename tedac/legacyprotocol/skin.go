package legacyprotocol

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Skin represents the skin of a player as sent over network. The skin holds a texture and a model, and
// optional animations which may be present when the skin is created using persona or bought from the
// marketplace.
type Skin struct {
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string
	// SkinResourcePatch is a JSON encoded object holding some fields that point to the geometry that the
	// skin has.
	// The JSON object that this holds specifies the way that the geometry of animations and the default skin
	// of the player are combined.
	SkinResourcePatch []byte
	// SkinImageWidth and SkinImageHeight hold the dimensions of the skin image. Note that these are not the
	// dimensions in bytes, but in pixels.
	SkinImageWidth, SkinImageHeight uint32
	// SkinData is a byte slice of SkinImageWidth * SkinImageHeight bytes. It is an RGBA ordered byte
	// representation of the skin pixels.
	SkinData []byte
	// Animations is a list of all animations that the skin has.
	Animations []protocol.SkinAnimation
	// CapeImageWidth and CapeImageHeight hold the dimensions of the cape image. Note that these are not the
	// dimensions in bytes, but in pixels.
	CapeImageWidth, CapeImageHeight uint32
	// CapeData is a byte slice of 64*32*4 bytes. It is a RGBA ordered byte representation of the cape
	// colours, much like the SkinData.
	CapeData []byte
	// SkinGeometry is a JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry []byte
	// TODO: Find out what value AnimationData holds and when it does hold something.
	AnimationData []byte
	// PremiumSkin specifies if this is a skin that was purchased from the marketplace.
	PremiumSkin bool
	// PersonaSkin specifies if this is a skin that was created using the in-game skin creator.
	PersonaSkin bool
	// PersonaCapeOnClassicSkin specifies if the skin had a Persona cape (in-game skin creator cape) equipped
	// on a classic skin.
	PersonaCapeOnClassicSkin bool
	// CapeID is a unique identifier that identifies the cape. It usually holds a UUID in it.
	CapeID string
	// FullSkinID is an ID that represents the skin in full. The actual functionality is unknown: The client
	// does not seem to send a value for this.
	FullSkinID string
	// SkinColour is a hex representation (including #) of the base colour of the skin. An example of the
	// colour sent here is '#b37b62'.
	SkinColour string
	// ArmSize is the size of the arms of the player's model. This is either 'wide' (generally for male skins)
	// or 'slim' (generally for female skins).
	ArmSize string
	// PersonaPieces is a list of all persona pieces that the skin is composed of.
	PersonaPieces []protocol.PersonaPiece
	// PieceTintColours is a list of specific tint colours for (some of) the persona pieces found in the list
	// above.
	PieceTintColours []protocol.PersonaPieceTintColour
	// Trusted specifies if the skin is 'trusted'. No code should rely on this field, as any proxy or client
	// can easily change it.
	Trusted bool
}

// WriteSerialisedSkin writes a Skin x to Writer w. WriteSerialisedSkin panics if the fields of the skin
// have invalid values, usually indicating that the dimensions of the skin images are incorrect.
func WriteSerialisedSkin(w protocol.IO, x *Skin) {
	if err := x.validate(); err != nil {
		panic(err)
	}
	w.String(&x.SkinID)
	w.ByteSlice(&x.SkinResourcePatch)
	w.Uint32(&x.SkinImageWidth)
	w.Uint32(&x.SkinImageHeight)
	w.ByteSlice(&x.SkinData)
	l := uint32(len(x.Animations))
	w.Uint32(&l)
	for _, anim := range x.Animations {
		anim.Marshal(w)
	}
	w.Uint32(&x.CapeImageWidth)
	w.Uint32(&x.CapeImageHeight)
	w.ByteSlice(&x.CapeData)
	w.ByteSlice(&x.SkinGeometry)
	w.ByteSlice(&x.AnimationData)
	w.Bool(&x.PremiumSkin)
	w.Bool(&x.PersonaSkin)
	w.Bool(&x.PersonaCapeOnClassicSkin)
	w.String(&x.CapeID)
	w.String(&x.FullSkinID)
	w.String(&x.ArmSize)
	w.String(&x.SkinColour)
	l = uint32(len(x.PersonaPieces))
	w.Uint32(&l)
	for _, piece := range x.PersonaPieces {
		piece.Marshal(w)
	}
	l = uint32(len(x.PieceTintColours))
	w.Uint32(&l)
	for _, tint := range x.PieceTintColours {
		tint.Marshal(w)
	}
}

// SerialisedSkin reads a Skin x from Reader r.
func SerialisedSkin(r *protocol.Reader, x *Skin) {
	var animationCount, count uint32

	r.String(&x.SkinID)
	r.ByteSlice(&x.SkinResourcePatch)
	r.Uint32(&x.SkinImageWidth)
	r.Uint32(&x.SkinImageHeight)
	r.ByteSlice(&x.SkinData)
	r.Uint32(&animationCount)

	x.Animations = make([]protocol.SkinAnimation, animationCount)
	for i := uint32(0); i < animationCount; i++ {
		x.Animations[i].Marshal(r)
	}
	r.Uint32(&x.CapeImageWidth)
	r.Uint32(&x.CapeImageHeight)
	r.ByteSlice(&x.CapeData)
	r.ByteSlice(&x.SkinGeometry)
	r.ByteSlice(&x.AnimationData)
	r.Bool(&x.PremiumSkin)
	r.Bool(&x.PersonaSkin)
	r.Bool(&x.PersonaCapeOnClassicSkin)
	r.String(&x.CapeID)
	r.String(&x.FullSkinID)
	r.String(&x.ArmSize)
	r.String(&x.SkinColour)

	r.Uint32(&count)
	x.PersonaPieces = make([]protocol.PersonaPiece, count)
	for i := uint32(0); i < count; i++ {
		x.PersonaPieces[i].Marshal(r)
	}
	r.Uint32(&count)
	x.PieceTintColours = make([]protocol.PersonaPieceTintColour, count)
	for i := uint32(0); i < count; i++ {
		x.PieceTintColours[i].Marshal(r)
	}
	if err := x.validate(); err != nil {
		r.InvalidValue(fmt.Sprintf("Skin %v", x.SkinID), "serialised skin", err.Error())
	}
}

// validate checks the skin and makes sure every one of its values are correct. It checks the image dimensions
// and makes sure they match the image size of the skin, cape and the skin's animations.
func (skin Skin) validate() error {
	if skin.SkinImageHeight*skin.SkinImageWidth*4 != uint32(len(skin.SkinData)) {
		return fmt.Errorf("expected size of skin is %vx%v (%v bytes total), but got %v bytes", skin.SkinImageWidth, skin.SkinImageHeight, skin.SkinImageHeight*skin.SkinImageWidth*4, len(skin.SkinData))
	}
	if skin.CapeImageHeight*skin.CapeImageWidth*4 != uint32(len(skin.CapeData)) {
		return fmt.Errorf("expected size of cape is %vx%v (%v bytes total), but got %v bytes", skin.CapeImageWidth, skin.CapeImageHeight, skin.CapeImageHeight*skin.CapeImageWidth*4, len(skin.CapeData))
	}
	for i, animation := range skin.Animations {
		if animation.ImageHeight*animation.ImageWidth*4 != uint32(len(animation.ImageData)) {
			return fmt.Errorf("expected size of animation %v is %vx%v (%v bytes total), but got %v bytes", i, animation.ImageWidth, animation.ImageHeight, animation.ImageHeight*animation.ImageWidth*4, len(animation.ImageData))
		}
	}
	return nil
}

// LegacySkin ...
func LegacySkin(skin protocol.Skin) Skin {
	return Skin{
		SkinID:                   skin.SkinID,
		SkinResourcePatch:        skin.SkinResourcePatch,
		SkinImageWidth:           skin.SkinImageWidth,
		SkinImageHeight:          skin.SkinImageHeight,
		SkinData:                 skin.SkinData,
		Animations:               skin.Animations,
		CapeImageWidth:           skin.CapeImageWidth,
		CapeImageHeight:          skin.CapeImageHeight,
		CapeData:                 skin.CapeData,
		SkinGeometry:             skin.SkinGeometry,
		AnimationData:            skin.AnimationData,
		PremiumSkin:              skin.PremiumSkin,
		PersonaSkin:              skin.PersonaSkin,
		PersonaCapeOnClassicSkin: skin.PersonaCapeOnClassicSkin,
		CapeID:                   skin.CapeID,
		FullSkinID:               skin.FullID,
		SkinColour:               skin.SkinColour,
		ArmSize:                  skin.ArmSize,
		PersonaPieces:            skin.PersonaPieces,
		PieceTintColours:         skin.PieceTintColours,
		Trusted:                  skin.Trusted,
	}
}

// LatestSkin ...
func LatestSkin(skin Skin) protocol.Skin {
	return protocol.Skin{
		SkinID:                    skin.SkinID,
		SkinResourcePatch:         skin.SkinResourcePatch,
		SkinImageWidth:            skin.SkinImageWidth,
		SkinImageHeight:           skin.SkinImageHeight,
		SkinData:                  skin.SkinData,
		Animations:                skin.Animations,
		CapeImageWidth:            skin.CapeImageWidth,
		CapeImageHeight:           skin.CapeImageHeight,
		CapeData:                  skin.CapeData,
		SkinGeometry:              skin.SkinGeometry,
		AnimationData:             skin.AnimationData,
		GeometryDataEngineVersion: []byte(protocol.CurrentVersion),
		PremiumSkin:               skin.PremiumSkin,
		PersonaSkin:               skin.PersonaSkin,
		PersonaCapeOnClassicSkin:  skin.PersonaCapeOnClassicSkin,
		CapeID:                    skin.CapeID,
		FullID:                    skin.FullSkinID,
		SkinColour:                skin.SkinColour,
		ArmSize:                   skin.ArmSize,
		PersonaPieces:             skin.PersonaPieces,
		PieceTintColours:          skin.PieceTintColours,
		Trusted:                   skin.Trusted,
	}
}
