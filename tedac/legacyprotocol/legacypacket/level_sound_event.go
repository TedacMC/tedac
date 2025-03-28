package legacypacket

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// LevelSoundEvent is sent by the server to make any kind of built-in sound heard to a player. It is sent to,
// for example, play a stepping sound or a shear sound. The packet is also sent by the client, in which case
// it could be forwarded by the server to the other players online. If possible, the packets from the client
// should be ignored however, and the server should play them on its own accord.
type LevelSoundEvent struct {
	// SoundType is the type of the sound to play. It is one of the constants above. Some of the sound types
	// require additional data, which is set in the EventData field.
	SoundType uint32
	// Position is the position of the sound event. The player will be able to hear the direction of the sound
	// based on what position is sent here.
	Position mgl32.Vec3
	// ExtraData is a packed integer that some sound types use to provide extra data. An example of this is
	// the note sound, which is composed of a pitch and an instrument type.
	ExtraData int32
	// EntityType is the string entity type of the entity that emitted the sound, for example
	// 'minecraft:skeleton'. Some sound types use this entity type for additional data.
	EntityType string
	// BabyMob specifies if the sound should be that of a baby mob. It is most notably used for parrot
	// imitations, which will change based on if this field is set to true or not.
	BabyMob bool
	// DisableRelativeVolume specifies if the sound should be played relatively or not. If set to true, the
	// sound will have full volume, regardless of where the Position is, whereas if set to false, the sound's
	// volume will be based on the distance to Position.
	DisableRelativeVolume bool
}

// ID ...
func (*LevelSoundEvent) ID() uint32 {
	return packet.IDLevelSoundEvent
}

func (pk *LevelSoundEvent) Marshal(io protocol.IO) {
	io.Varuint32(&pk.SoundType)
	io.Vec3(&pk.Position)
	io.Varint32(&pk.ExtraData)
	io.String(&pk.EntityType)
	io.Bool(&pk.BabyMob)
	io.Bool(&pk.DisableRelativeVolume)
}
