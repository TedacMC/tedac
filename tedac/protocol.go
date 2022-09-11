package tedac

import (
	"crypto/aes"

	legacypacket "github.com/tedacmc/tedac/tedac/packet"
	"github.com/tedacmc/tedac/tedac/raknet"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Protocol represents the v1.12.0 Protocol implementation.
type Protocol struct{}

// ID ...
func (Protocol) ID() int32 {
	return 361
}

// Ver ...
func (Protocol) Ver() string {
	return "1.12.0"
}

// Packets ...
func (Protocol) Packets() packet.Pool {
	p := packet.NewPool()
	p[packet.IDMovePlayer] = func() packet.Packet { return &legacypacket.MovePlayer{} }
	return packet.NewPool()
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	block, _ := aes.NewCipher(key[:])
	return raknet.NewCfb(
		key[:],
		block,
		append([]byte(nil), key[:aes.BlockSize]...),
	)
}

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *legacypacket.MovePlayer:
		return []packet.Packet{
			&packet.MovePlayer{
				EntityRuntimeID:       pk.EntityRuntimeID,
				Position:              pk.Position,
				Pitch:                 pk.Pitch,
				Yaw:                   pk.Yaw,
				HeadYaw:               pk.HeadYaw,
				Mode:                  pk.Mode,
				OnGround:              pk.OnGround,
				RiddenEntityRuntimeID: pk.RiddenEntityRuntimeID,
				TeleportCause:         pk.TeleportCause,
			},
		}
	}
	return []packet.Packet{pk}
}

// ConvertFromLatest ...
func (Protocol) ConvertFromLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.MovePlayer:
		return []packet.Packet{
			&legacypacket.MovePlayer{
				EntityRuntimeID:       pk.EntityRuntimeID,
				Position:              pk.Position,
				Pitch:                 pk.Pitch,
				Yaw:                   pk.Yaw,
				HeadYaw:               pk.HeadYaw,
				Mode:                  pk.Mode,
				OnGround:              pk.OnGround,
				RiddenEntityRuntimeID: pk.RiddenEntityRuntimeID,
				TeleportCause:         pk.TeleportCause,
				// TeleportItem: ???
			},
		}
	}
	return []packet.Packet{pk}
}
