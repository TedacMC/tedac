package tedac

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacymappings"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
	legacypacket2 "github.com/tedacmc/tedac/tedac/legacyprotocol/legacypacket"
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
	p[packet.IDMovePlayer] = func() packet.Packet { return &legacypacket2.MovePlayer{} }
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
	case *legacypacket2.MovePlayer:
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
	case *legacypacket2.PlayerAction:
		return []packet.Packet{
			&packet.PlayerAction{
				EntityRuntimeID: pk.EntityRuntimeID,
				ActionType:      pk.ActionType,
				BlockPosition:   pk.BlockPosition,
				BlockFace:       pk.BlockFace,
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
			&legacypacket2.MovePlayer{
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
	case *packet.StartGame:
		return []packet.Packet{
			&legacypacket2.StartGame{
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
				RainLevel:                      pk.RainLevel,
				LightningLevel:                 pk.LightningLevel,
				ConfirmedPlatformLockedContent: pk.ConfirmedPlatformLockedContent,
				MultiPlayerGame:                pk.MultiPlayerGame,
				LANBroadcastEnabled:            pk.LANBroadcastEnabled,
				XBLBroadcastMode:               pk.XBLBroadcastMode,
				PlatformBroadcastMode:          pk.PlatformBroadcastMode,
				CommandsEnabled:                pk.CommandsEnabled,
				TexturePackRequired:            pk.TexturePackRequired,
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
				Trial:                          pk.Trial,
				Time:                           pk.Time,
				EnchantmentSeed:                pk.EnchantmentSeed,
				Blocks:                         legacymappings.Blocks(),
				Items:                          legacymappings.Items(),
				MultiPlayerCorrelationID:       pk.MultiPlayerCorrelationID,
				GameRules: lo.SliceToMap(pk.GameRules, func(rule protocol.GameRule) (string, any) {
					return rule.Name, rule.Value
				}),
			},
		}
	// case *packet.AvailableActorIdentifiers, *packet.CraftingData, *packet.UpdateAttributes, *packet.SetActorData, *packet.UpdateAbilities, *packet.UpdateAdventureSettings, *packet.CreativeContent, *packet.LevelChunk:
	// 	// TODO: Properly handle these!
	// 	return nil
	case *packet.ActorPickRequest:
		return []packet.Packet{
			&legacypacket2.ActorPickRequest{
				EntityUniqueID: pk.EntityUniqueID,
				HotBarSlot:     pk.HotBarSlot,
			},
		}
	case *packet.AddActor:
		attributes := []legacyprotocol.Attribute{}
		for _, a := range pk.Attributes {
			attributes = append(attributes, legacyprotocol.Attribute{
				Name:  a.Name,
				Value: a.Value,
				Max:   a.Max,
				Min:   a.Min,
			})
		}
		links := []legacyprotocol.EntityLink{}
		for _, l := range pk.EntityLinks {
			links = append(links, legacyprotocol.EntityLink{
				RiddenEntityUniqueID: l.RiddenEntityUniqueID,
				RiderEntityUniqueID:  l.RiddenEntityUniqueID,
				Type:                 l.Type,
				Immediate:            l.Immediate,
			})
		}
		return []packet.Packet{
			&legacypacket2.AddActor{
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityType:      pk.EntityType,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Pitch:           pk.Pitch,
				Yaw:             pk.Yaw,
				HeadYaw:         pk.HeadYaw,
				Attributes:      attributes,
				EntityMetadata:  pk.EntityMetadata,
				EntityLinks:     links,
			},
		}
	case *packet.AddPlayer:
		links := []legacyprotocol.EntityLink{}
		for _, l := range pk.EntityLinks {
			links = append(links, legacyprotocol.EntityLink{
				RiddenEntityUniqueID: l.RiddenEntityUniqueID,
				RiderEntityUniqueID:  l.RiddenEntityUniqueID,
				Type:                 l.Type,
				Immediate:            l.Immediate,
			})
		}
		return []packet.Packet{
			&legacypacket2.AddPlayer{
				UUID:            pk.UUID,
				Username:        pk.Username,
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				PlatformChatID:  pk.PlatformChatID,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Pitch:           pk.Pitch,
				Yaw:             pk.Yaw,
				HeadYaw:         pk.HeadYaw,
				// HeldItem: pk.HeldItem, ???
				EntityMetadata:         pk.EntityMetadata,
				CommandPermissionLevel: uint32(pk.CommandPermissions),
				PermissionLevel:        uint32(pk.PlayerPermissions),
				EntityLinks:            links,
				DeviceID:               pk.DeviceID,
			},
		}
	case *packet.AvailableActorIdentifiers:
		pk.SerialisedEntityIdentifiers = []byte(legacypacket2.AaiNiggerHardcode)
		return []packet.Packet{pk}
	case *packet.BiomeDefinitionList:
		pk.SerialisedBiomeDefinitions = []byte(legacypacket2.BdlNiggerHardcode)
		return []packet.Packet{pk}
	case *packet.ContainerClose:
		return []packet.Packet{
			&legacypacket2.ContainerClose{
				WindowID: pk.WindowID,
			},
		}
	}
	fmt.Printf("1.19 -> 1.12: %T\n", pk)
	return []packet.Packet{pk}
}
