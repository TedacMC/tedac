package legacypacket

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	InputFlagAscend = 1 << iota
	InputFlagDescend
	InputFlagNorthJump
	InputFlagJumpDown
	InputFlagSprintDown
	InputFlagChangeHeight
	InputFlagJumping
	InputFlagAutoJumpingInWater
	InputFlagSneaking
	InputFlagSneakDown
	InputFlagUp
	InputFlagDown
	InputFlagLeft
	InputFlagRight
	InputFlagUpLeft
	InputFlagUpRight
	InputFlagWantUp
	InputFlagWantDown
	InputFlagWantDownSlow
	InputFlagWantUpSlow
	InputFlagSprinting
	InputFlagAscendScaffolding
	InputFlagDescendScaffolding
	InputFlagSneakToggleDown
	InputFlagPersistSneak
)

const (
	InputModeMouse = iota + 1
	InputModeTouch
	InputModeGamePad
	InputModeMotionController
)

const (
	PlayModeNormal = iota
	PlayModeTeaser
	PlayModeScreen
	PlayModeViewer
	PlayModeReality
	PlayModePlacement
	PlayModeLivingRoom
	PlayModeExitLevel
	PlayModeExitLevelLivingRoom
	PlayModeNumModes
)

// PlayerAuthInput is sent by the client to allow for server authoritative movement. It is used to synchronise
// the player input with the position server-side.
// The client sends this packet when the ServerAuthoritativeMovementMode field in the StartGame packet is set
// to true, instead of the MovePlayer packet. The client will send this packet once every tick.
type PlayerAuthInput struct {
	// Pitch and Yaw hold the rotation that the player reports it has.
	Pitch, Yaw float32
	// Position holds the position that the player reports it has.
	Position mgl32.Vec3
	// MoveVector is a Vec2 that specifies the direction in which the player moved, as a combination of X/Z
	// values which are created using the WASD/controller stick state.
	MoveVector mgl32.Vec2
	// HeadYaw is the horizontal rotation of the head that the player reports it has.
	HeadYaw float32
	// InputData is a combination of bit flags that together specify the way the player moved last tick. It
	// is a combination of the flags above.
	InputData uint64
	// InputMode specifies the way that the client inputs data to the screen. It is one of the constants that
	// may be found above.
	InputMode uint32
	// PlayMode specifies the way that the player is playing. The values it holds, which are rather random,
	// may be found above.
	PlayMode uint32
	// GazeDirection is the direction in which the player is gazing, when the PlayMode is PlayModeReality: In
	// other words, when the player is playing in virtual reality.
	GazeDirection mgl32.Vec3
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
	// Delta was the delta between the old and the new position. There isn't any practical use for this field
	// as it can be calculated by the server itself.
	Delta mgl32.Vec3
}

// ID ...
func (pk *PlayerAuthInput) ID() uint32 {
	return packet.IDPlayerAuthInput
}

// Marshal ...
func (pk *PlayerAuthInput) Marshal(w protocol.IO) {
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Vec3(&pk.Position)
	w.Vec2(&pk.MoveVector)
	w.Float32(&pk.HeadYaw)
	w.Varuint64(&pk.InputData)
	w.Varuint32(&pk.InputMode)
	w.Varuint32(&pk.PlayMode)
	if pk.PlayMode == PlayModeReality {
		w.Vec3(&pk.GazeDirection)
	}
	w.Varuint64(&pk.Tick)
	w.Vec3(&pk.Delta)
}
