package tedac

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/chunk"
	"github.com/tedacmc/tedac/tedac/latestmappings"
	"github.com/tedacmc/tedac/tedac/legacychunk"
	"github.com/tedacmc/tedac/tedac/legacymappings"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
	"github.com/tedacmc/tedac/tedac/legacyprotocol/legacypacket"
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
	return "1.12.1"
}

// Packets ...
func (Protocol) Packets() packet.Pool {
	pool := packet.NewPool()
	pool[packet.IDActorPickRequest] = func() packet.Packet { return &legacypacket.ActorPickRequest{} }
	pool[packet.IDAddActor] = func() packet.Packet { return &legacypacket.AddActor{} }
	pool[packet.IDAddPlayer] = func() packet.Packet { return &legacypacket.AddPlayer{} }
	pool[packet.IDAvailableActorIdentifiers] = func() packet.Packet { return &legacypacket.AvailableActorIdentifiers{} }
	//pool[packet.IDAvailableCommands] = func() packet.Packet { return &legacypacket.AvailableCommands{} }
	pool[packet.IDBiomeDefinitionList] = func() packet.Packet { return &legacypacket.BiomeDefinitionList{} }
	pool[packet.IDContainerClose] = func() packet.Packet { return &legacypacket.ContainerClose{} }
	//pool[packet.IDCraftingData] = func() packet.Packet { return &legacypacket.CraftingData{} }
	pool[packet.IDLevelChunk] = func() packet.Packet { return &legacypacket.LevelChunk{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	pool[packet.IDMovePlayer] = func() packet.Packet { return &legacypacket.MovePlayer{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	pool[packet.IDPlayerList] = func() packet.Packet { return &legacypacket.PlayerList{} }
	pool[packet.IDPlayerSkin] = func() packet.Packet { return &legacypacket.PlayerSkin{} }
	pool[packet.IDSetActorData] = func() packet.Packet { return &legacypacket.SetActorData{} }
	pool[packet.IDStartGame] = func() packet.Packet { return &legacypacket.StartGame{} }
	pool[packet.IDText] = func() packet.Packet { return &legacypacket.Text{} }
	pool[packet.IDUpdateAttributes] = func() packet.Packet { return &legacypacket.UpdateAttributes{} }
	return pool
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return newCFBEncryption(key[:])
}

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	// fmt.Printf("1.12 -> 1.19: %T\n", pk)
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
	case *legacypacket.PlayerAction:
		return []packet.Packet{
			&packet.PlayerAction{
				EntityRuntimeID: pk.EntityRuntimeID,
				ActionType:      pk.ActionType,
				BlockPosition:   pk.BlockPosition,
				BlockFace:       pk.BlockFace,
			},
		}
	case *legacypacket.ModalFormResponse:
		return []packet.Packet{
			&packet.ModalFormResponse{
				FormID:       pk.FormID,
				ResponseData: protocol.Option(pk.ResponseData),
				CancelReason: protocol.Option[uint8](packet.ModalFormCancelReasonUserClosed), // idfk man im not payed enough
			},
		}
	}
	return []packet.Packet{pk}
}

// ConvertFromLatest ...
func (Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	switch pk := pk.(type) {
	case *packet.StartGame:
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
	case *packet.LevelChunk:
		if pk.SubChunkRequestMode != protocol.SubChunkRequestModeLegacy {
			// TODO: Support other sub-chunk request modes.
			return nil
		}

		buf := bytes.NewBuffer(pk.RawPayload)
		oldFormat := conn.GameData().BaseGameVersion == "1.17.40"
		c, err := chunk.NetworkDecode(latestAirRID, buf, int(pk.SubChunkCount), oldFormat, world.Overworld.Range())
		if err != nil {
			fmt.Println(err)
			return nil
		}

		writeBuf, data := bytes.NewBuffer(nil), legacychunk.Encode(downgradeChunk(c, oldFormat), legacychunk.NetworkEncoding)
		for i := range data.SubChunks {
			_, _ = writeBuf.Write(data.SubChunks[i])
		}
		_, _ = writeBuf.Write(data.Data2D)

		return []packet.Packet{
			&legacypacket.LevelChunk{
				BlobHashes:    pk.BlobHashes,
				CacheEnabled:  pk.CacheEnabled,
				Position:      pk.Position,
				RawPayload:    append(writeBuf.Bytes(), buf.Bytes()...),
				SubChunkCount: uint32(len(data.SubChunks)),
			},
		}
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
			},
		}
	case *packet.ActorPickRequest:
		return []packet.Packet{
			&legacypacket.ActorPickRequest{
				EntityUniqueID: pk.EntityUniqueID,
				HotBarSlot:     pk.HotBarSlot,
			},
		}
	case *packet.AddActor:
		var attributes []legacyprotocol.Attribute
		for _, a := range pk.Attributes {
			attributes = append(attributes, legacyprotocol.Attribute{
				Name:  a.Name,
				Value: a.Value,
				Max:   a.Max,
				Min:   a.Min,
			})
		}
		var links []legacyprotocol.EntityLink
		for _, l := range pk.EntityLinks {
			links = append(links, legacyprotocol.EntityLink{
				RiddenEntityUniqueID: l.RiddenEntityUniqueID,
				RiderEntityUniqueID:  l.RiddenEntityUniqueID,
				Type:                 l.Type,
				Immediate:            l.Immediate,
			})
		}
		return []packet.Packet{
			&legacypacket.AddActor{
				Attributes:      attributes,
				EntityLinks:     links,
				EntityMetadata:  pk.EntityMetadata,
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityType:      pk.EntityType,
				EntityUniqueID:  pk.EntityUniqueID,
				HeadYaw:         pk.HeadYaw,
				Pitch:           pk.Pitch,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Yaw:             pk.Yaw,
			},
		}
	case *packet.AddPlayer:
		var links []legacyprotocol.EntityLink
		for _, l := range pk.EntityLinks {
			links = append(links, legacyprotocol.EntityLink{
				RiddenEntityUniqueID: l.RiddenEntityUniqueID,
				RiderEntityUniqueID:  l.RiddenEntityUniqueID,
				Type:                 l.Type,
				Immediate:            l.Immediate,
			})
		}
		return []packet.Packet{
			&legacypacket.AddPlayer{
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
		pk.SerialisedEntityIdentifiers = []byte(legacypacket.AaiNiggerHardcode)
		return []packet.Packet{pk}
	case *packet.BiomeDefinitionList:
		pk.SerialisedBiomeDefinitions = []byte(legacypacket.BdlNiggerHardcode)
		return []packet.Packet{pk}
	case *packet.ContainerClose:
		return []packet.Packet{
			&legacypacket.ContainerClose{
				WindowID: pk.WindowID,
			},
		}
	case *packet.PlayerList:
		var entries []legacypacket.PlayerListEntry
		for _, entry := range pk.Entries {
			var patch struct {
				Geometry struct {
					Default string
				}
			}
			err := json.Unmarshal(entry.Skin.SkinResourcePatch, &patch)
			if err != nil {
				fmt.Println(err)
				continue
			}
			entries = append(entries, legacypacket.PlayerListEntry{
				UUID:             entry.UUID,
				EntityUniqueID:   entry.EntityUniqueID,
				Username:         entry.Username,
				SkinID:           entry.Skin.SkinID,
				SkinData:         entry.Skin.SkinData,
				CapeData:         entry.Skin.CapeData,
				SkinGeometryName: patch.Geometry.Default,
				SkinGeometry:     entry.Skin.SkinGeometry,
				PlatformChatID:   entry.PlatformChatID,
				XUID:             entry.XUID,
			})
		}
		return []packet.Packet{
			&legacypacket.PlayerList{
				ActionType: pk.ActionType,
				Entries:    entries,
			},
		}
	case *packet.UpdateAttributes:
		var attributes []legacyprotocol.Attribute
		for _, a := range pk.Attributes {
			attributes = append(attributes, legacyprotocol.Attribute{
				Name:  a.Name,
				Value: a.Value,
				Max:   a.Max,
				Min:   a.Min,
			})
		}
		return []packet.Packet{
			&legacypacket.UpdateAttributes{
				EntityRuntimeID: pk.EntityRuntimeID,
				Attributes:      attributes,
			},
		}
	case *packet.SetActorData:
		return []packet.Packet{
			&legacypacket.SetActorData{
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityMetadata:  pk.EntityMetadata,
			},
		}
	case *packet.PlayerSkin:
		var patch struct {
			Geometry struct {
				Default string
			}
		}
		err := json.Unmarshal(pk.Skin.SkinResourcePatch, &patch)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return []packet.Packet{
			&legacypacket.PlayerSkin{
				UUID:             pk.UUID,
				SkinID:           pk.Skin.SkinID,
				NewSkinName:      pk.NewSkinName,
				OldSkinName:      pk.OldSkinName,
				SkinData:         pk.Skin.SkinData,
				CapeData:         pk.Skin.CapeData,
				SkinGeometryName: patch.Geometry.Default,
				SkinGeometry:     pk.Skin.SkinGeometry,
				PremiumSkin:      pk.Skin.PremiumSkin,
			},
		}
	}
	//fmt.Printf("1.19 -> 1.12: %T\n", pk)
	return []packet.Packet{pk}
}

var (
	// latestAirRID is the runtime ID of the air block in the latest version of the game.
	latestAirRID, _ = latestmappings.StateToRuntimeID("minecraft:air", nil)
	// legacyAirRID is the runtime ID of the air block in the v1.12.0 version.
	legacyAirRID, _ = legacymappings.StateToRuntimeID("minecraft:air", nil)
)

// downgradeChunk downgrades a chunk from the latest version to the v1.12.0 equivalent.
func downgradeChunk(chunk *chunk.Chunk, oldFormat bool) *legacychunk.Chunk {
	// First downgrade the blocks.
	subs := chunk.Sub()
	if oldFormat {
		subs = subs[:len(subs)-4]
	} else {
		subs = subs[4 : len(subs)-4]
	}

	downgraded := legacychunk.New(legacyAirRID)
	for subInd, sub := range subs {
		for layerInd, layer := range sub.Layers() {
			downgradedLayer := downgraded.Sub()[subInd].Layer(uint8(layerInd))
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					for y := uint8(0); y < 16; y++ {
						latestRuntimeID := layer.At(x, y, z)
						if latestRuntimeID == latestAirRID {
							// Don't bother with air.
							continue
						}

						name, properties, ok := latestmappings.RuntimeIDToState(latestRuntimeID)
						if !ok {
							// Unknown runtime ID, ignore this position.
							continue
						}
						rid, ok := legacymappings.StateToRuntimeID(name, properties)
						if !ok {
							// Unknown state, ignore this position.
							continue
						}

						downgradedLayer.SetRuntimeID(x, y, z, rid)
					}
				}
			}
		}
	}

	// Then downgrade the biomes.
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			// Use the highest block as an estimate for the biome, since we only have 2D biomes.
			downgraded.SetBiomeID(x, z, uint8(chunk.Biome(x, chunk.HighestBlock(x, z), z)))
		}
	}
	return downgraded
}
