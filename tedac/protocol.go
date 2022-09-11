package tedac

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	legacypacket "github.com/tedacmc/tedac/tedac/packet"
	_ "github.com/tedacmc/tedac/tedac/raknet"
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
	return newCFBEncryption(key[:])
}

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	fmt.Printf("1.12 -> 1.19: %T\n", pk)
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
	fmt.Printf("1.19 -> 1.12: %T\n", pk)
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
	case *packet.StartGame:
		gamerules := make(map[string]interface{})
		for _, gr := range pk.GameRules {
			gamerules[gr.Name] = gr.Value
		}
		return []packet.Packet{
			&legacypacket.StartGame{
				EntityUniqueID:                 pk.EntityUniqueID,
				EntityRuntimeID:                pk.EntityRuntimeID,
				PlayerGameMode:                 pk.PlayerGameMode,
				PlayerPosition:                 pk.PlayerPosition,
				Pitch:                          pk.Pitch,
				Yaw:                            pk.Yaw,
				WorldSeed:                      int32(pk.WorldSeed),
				Dimension:                      pk.Dimension,
				Generator:                      pk.Generator,
				WorldGameMode:                  pk.WorldGameMode,
				Difficulty:                     pk.Difficulty,
				WorldSpawn:                     pk.WorldSpawn,
				AchievementsDisabled:           pk.AchievementsDisabled,
				DayCycleLockTime:               pk.DayCycleLockTime,
				EducationMode:                  false, // no nigga uses this
				EducationFeaturesEnabled:       false, // again, no nigga uses this
				RainLevel:                      pk.RainLevel,
				LightningLevel:                 pk.LightningLevel,
				ConfirmedPlatformLockedContent: pk.ConfirmedPlatformLockedContent,
				MultiPlayerGame:                pk.MultiPlayerGame,
				LANBroadcastEnabled:            pk.LANBroadcastEnabled,
				XBLBroadcastMode:               pk.XBLBroadcastMode,
				PlatformBroadcastMode:          pk.PlatformBroadcastMode,
				CommandsEnabled:                pk.CommandsEnabled,
				TexturePackRequired:            pk.TexturePackRequired,
				GameRules:                      gamerules,
				BonusChestEnabled:              pk.BonusChestEnabled,
				StartWithMapEnabled:            pk.StartWithMapEnabled,
				PlayerPermissions:              int32(pk.PlayerPermissions),
				ServerChunkTickRadius:          pk.ServerChunkTickRadius,
				HasLockedBehaviourPack:         pk.HasLockedBehaviourPack,
				HasLockedTexturePack:           pk.HasLockedTexturePack,
				FromLockedWorldTemplate:        pk.FromLockedWorldTemplate,
				MSAGamerTagsOnly:               pk.MSAGamerTagsOnly,
				FromWorldTemplate:              pk.FromWorldTemplate,
				WorldTemplateSettingsLocked:    pk.WorldTemplateSettingsLocked,
				OnlySpawnV1Villagers:           pk.OnlySpawnV1Villagers,
				LevelID:                        pk.LevelID,
				WorldName:                      pk.WorldName,
				PremiumWorldTemplateID:         "",
				Trial:                          pk.Trial,
				Time:                           pk.Time,
				EnchantmentSeed:                pk.EnchantmentSeed,
				Blocks:                         pk.Blocks,
				Items:                          pk.Items,
				MultiPlayerCorrelationID:       pk.MultiPlayerCorrelationID,
			},
		}
	}
	return []packet.Packet{pk}
}
