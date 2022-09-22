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
	//pool[packet.IDContainerClose] = func() packet.Packet { return &legacypacket.ContainerClose{} }
	pool[packet.IDInventoryTransaction] = func() packet.Packet { return &legacypacket.InventoryTransaction{} }
	pool[packet.IDMobEquipment] = func() packet.Packet { return &legacypacket.MobEquipment{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	pool[packet.IDMovePlayer] = func() packet.Packet { return &legacypacket.MovePlayer{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	return pool
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return newCFBEncryption(key[:])
}

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	//fmt.Printf("1.12 -> 1.19.30: %T\n", pk)
	switch pk := pk.(type) {
	case *legacypacket.MovePlayer:
		if conn.GameData().PlayerMovementSettings.MovementType != protocol.PlayerMovementModeClient {
			return []packet.Packet{
				&packet.PlayerAuthInput{
					Pitch:    pk.Pitch,
					Yaw:      pk.Yaw,
					Position: pk.Position,
					HeadYaw:  pk.HeadYaw,
				},
			}
		}
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
	case *legacypacket.InventoryTransaction:
		actions := make([]protocol.InventoryAction, 0, len(pk.Actions))
		for _, action := range pk.Actions {
			actions = append(actions, protocol.InventoryAction{
				SourceType:    action.SourceType,
				WindowID:      action.WindowID,
				SourceFlags:   action.SourceFlags,
				InventorySlot: action.InventorySlot,
				OldItem:       protocol.ItemInstance{Stack: upgradeItem(action.OldItem)},
				NewItem:       protocol.ItemInstance{Stack: upgradeItem(action.NewItem)},
			})
		}

		var transactionData protocol.InventoryTransactionData
		switch data := pk.TransactionData.(type) {
		case *legacyprotocol.NormalTransactionData:
			transactionData = &protocol.NormalTransactionData{}
		case *legacyprotocol.MismatchTransactionData:
			transactionData = &protocol.MismatchTransactionData{}
		case *legacyprotocol.UseItemTransactionData:
			transactionData = &protocol.UseItemTransactionData{
				ActionType:      data.ActionType,
				BlockPosition:   data.BlockPosition,
				BlockFace:       data.BlockFace,
				HotBarSlot:      data.HotBarSlot,
				HeldItem:        protocol.ItemInstance{Stack: upgradeItem(data.HeldItem)},
				Position:        data.Position,
				ClickedPosition: data.ClickedPosition,
				BlockRuntimeID:  upgradeBlockRuntimeID(data.BlockRuntimeID),
			}
		case *legacyprotocol.UseItemOnEntityTransactionData:
			transactionData = &protocol.UseItemOnEntityTransactionData{
				TargetEntityRuntimeID: data.TargetEntityRuntimeID,
				ActionType:            data.ActionType,
				HotBarSlot:            data.HotBarSlot,
				HeldItem:              protocol.ItemInstance{Stack: upgradeItem(data.HeldItem)},
				Position:              data.Position,
				ClickedPosition:       data.ClickedPosition,
			}
		case *legacyprotocol.ReleaseItemTransactionData:
			transactionData = &protocol.ReleaseItemTransactionData{
				ActionType:   data.ActionType,
				HotBarSlot:   data.HotBarSlot,
				HeldItem:     protocol.ItemInstance{Stack: upgradeItem(data.HeldItem)},
				HeadPosition: data.HeadPosition,
			}
		}

		return []packet.Packet{
			&packet.InventoryTransaction{
				Actions:         actions,
				TransactionData: transactionData,
			},
		}
	case *legacypacket.ModalFormResponse:
		return []packet.Packet{
			&packet.ModalFormResponse{
				FormID:       pk.FormID,
				ResponseData: protocol.Option(pk.ResponseData),
			},
		}
	case *legacypacket.MobEquipment:
		return []packet.Packet{
			&packet.MobEquipment{
				EntityRuntimeID: pk.EntityRuntimeID,
				NewItem:         protocol.ItemInstance{Stack: upgradeItem(pk.NewItem)},
				InventorySlot:   pk.InventorySlot,
				HotBarSlot:      pk.HotBarSlot,
				WindowID:        pk.HotBarSlot,
			},
		}
	}
	return []packet.Packet{pk}
}

// ConvertFromLatest ...
func (Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	//fmt.Printf("1.19.30 -> 1.12: %T\n", pk)
	switch pk := pk.(type) {
	case *packet.StartGame:
		// Adjust our mappings to account for any possible custom blocks.
		latestmappings.Adjust(lo.Map(pk.Blocks, func(entry protocol.BlockEntry, _ int) latestmappings.State {
			return latestmappings.State{
				Name:    entry.Name,
				Version: chunk.CurrentBlockVersion,
			}
		}))

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

		writeBuf, data := bytes.NewBuffer(nil), legacychunk.Encode(downgradeChunk(c), legacychunk.NetworkEncoding)
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
	case *packet.UpdateBlock:
		pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
	case *packet.UpdateBlockSynced:
		pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
	case *packet.NetworkChunkPublisherUpdate:
		return []packet.Packet{
			&legacypacket.NetworkChunkPublisherUpdate{
				Position: pk.Position,
				Radius:   pk.Radius,
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
				UUID:                   pk.UUID,
				Username:               pk.Username,
				EntityUniqueID:         pk.AbilityData.EntityUniqueID,
				EntityRuntimeID:        pk.EntityRuntimeID,
				PlatformChatID:         pk.PlatformChatID,
				Position:               pk.Position,
				Velocity:               pk.Velocity,
				Pitch:                  pk.Pitch,
				Yaw:                    pk.Yaw,
				HeadYaw:                pk.HeadYaw,
				HeldItem:               downgradeItem(pk.HeldItem.Stack),
				EntityMetadata:         pk.EntityMetadata,
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				EntityLinks:            links,
				DeviceID:               pk.DeviceID,
			},
		}
	case *packet.MobEquipment:
		return []packet.Packet{
			&legacypacket.MobEquipment{
				EntityRuntimeID: pk.EntityRuntimeID,
				NewItem:         downgradeItem(pk.NewItem.Stack),
				InventorySlot:   pk.InventorySlot,
				HotBarSlot:      pk.HotBarSlot,
				WindowID:        pk.WindowID,
			},
		}
	case *packet.MobArmourEquipment:
		return []packet.Packet{
			&legacypacket.MobArmourEquipment{
				EntityRuntimeID: pk.EntityRuntimeID,
				Helmet:          downgradeItem(pk.Helmet.Stack),
				Chestplate:      downgradeItem(pk.Chestplate.Stack),
				Leggings:        downgradeItem(pk.Leggings.Stack),
				Boots:           downgradeItem(pk.Boots.Stack),
			},
		}
	case *packet.AddItemActor:
		return []packet.Packet{
			&legacypacket.AddItemActor{
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				Item:            downgradeItem(pk.Item.Stack),
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				EntityMetadata:  pk.EntityMetadata,
				FromFishing:     pk.FromFishing,
			},
		}
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
			_ = json.Unmarshal(entry.Skin.SkinResourcePatch, &patch)
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
	case *packet.UpdateAbilities:
		if len(pk.AbilityData.Layers) == 0 {
			// We need at least one layer.
			return nil
		}

		base, flags := pk.AbilityData.Layers[0].Values, uint32(0)
		flags &= ^uint32(packet.AdventureFlagWorldImmutable)

		if base&protocol.AbilityAttackPlayers != 0 {
			flags |= packet.AdventureSettingsFlagsNoPvM
		} else {
			flags &= ^uint32(packet.AdventureSettingsFlagsNoPvM)
		}

		flags &= ^uint32(packet.AdventureFlagAutoJump)

		if base&protocol.AbilityMayFly == 0 {
			flags |= packet.AdventureFlagAllowFlight
		} else {
			flags &= ^uint32(packet.AdventureFlagAllowFlight)
		}

		if base&protocol.AbilityNoClip == 0 {
			flags |= packet.AdventureFlagNoClip
		} else {
			flags &= ^uint32(packet.AdventureFlagNoClip)
		}

		if base&protocol.AbilityFlying == 0 {
			flags |= packet.AdventureFlagFlying
		} else {
			flags &= ^uint32(packet.AdventureFlagFlying)
		}

		return []packet.Packet{
			&packet.AdventureSettings{
				Flags:                  flags,
				PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
			},
		}
	case *packet.UpdateAttributes:
		return []packet.Packet{
			&legacypacket.UpdateAttributes{
				EntityRuntimeID: pk.EntityRuntimeID,
				Attributes: lo.Map(pk.Attributes, func(attribute protocol.Attribute, _ int) legacyprotocol.Attribute {
					return legacyprotocol.Attribute{
						Name:    attribute.Name,
						Value:   attribute.Value,
						Min:     attribute.Min,
						Max:     attribute.Max,
						Default: attribute.Default,
					}
				}),
			},
		}
	case *packet.SetActorData:
		return []packet.Packet{
			&legacypacket.SetActorData{
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityMetadata:  pk.EntityMetadata,
			},
		}
	case *packet.InventorySlot:
		return []packet.Packet{
			&legacypacket.InventorySlot{
				WindowID: pk.WindowID,
				Slot:     pk.Slot,
				NewItem:  downgradeItem(pk.NewItem.Stack),
			},
		}
	case *packet.InventoryContent:
		return []packet.Packet{
			&legacypacket.InventoryContent{
				WindowID: pk.WindowID,
				Content: lo.Map(pk.Content, func(instance protocol.ItemInstance, _ int) legacyprotocol.ItemStack {
					return downgradeItem(instance.Stack)
				}),
			},
		}
	case *packet.ResourcePacksInfo:
		return []packet.Packet{
			&legacypacket.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				BehaviourPacks: lo.Map(pk.BehaviourPacks, func(pack protocol.BehaviourPackInfo, _ int) legacyprotocol.ResourcePackInfo {
					return legacyprotocol.ResourcePackInfo{
						UUID:            pack.UUID,
						Version:         pack.Version,
						Size:            pack.Size,
						ContentKey:      pack.ContentKey,
						SubPackName:     pack.SubPackName,
						ContentIdentity: pack.ContentIdentity,
						HasScripts:      pack.HasScripts,
					}
				}),
				TexturePacks: lo.Map(pk.TexturePacks, func(pack protocol.TexturePackInfo, _ int) legacyprotocol.ResourcePackInfo {
					return legacyprotocol.ResourcePackInfo{
						UUID:            pack.UUID,
						Version:         pack.Version,
						Size:            pack.Size,
						ContentKey:      pack.ContentKey,
						SubPackName:     pack.SubPackName,
						ContentIdentity: pack.ContentIdentity,
						HasScripts:      pack.HasScripts,
					}
				}),
			},
		}
	case *packet.ResourcePackStack:
		return []packet.Packet{
			&legacypacket.ResourcePackStack{
				TexturePackRequired: pk.TexturePackRequired,
				BehaviourPacks:      pk.BehaviourPacks,
				TexturePacks:        pk.TexturePacks,
				Experimental:        len(pk.Experiments) > 0,
			},
		}
	case *packet.ResourcePackChunkData:
		return []packet.Packet{
			&legacypacket.ResourcePackChunkData{
				UUID:       pk.UUID,
				ChunkIndex: pk.ChunkIndex,
				DataOffset: pk.DataOffset,
				Data:       pk.Data,
			},
		}
	case *packet.LevelEvent:
		if pk.EventType == packet.LevelEventParticlesDestroyBlock || pk.EventType == packet.LevelEventParticlesCrackBlock {
			pk.EventData = int32(downgradeBlockRuntimeID(uint32(pk.EventData)))
		}
	case *packet.CreativeContent, *packet.AvailableCommands, *packet.ItemComponent:
		return nil
	case *packet.PlayerSkin:
		var patch struct {
			Geometry struct {
				Default string
			}
		}
		_ = json.Unmarshal(pk.Skin.SkinResourcePatch, &patch)
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
	return []packet.Packet{pk}
}

var (
	// latestAirRID is the runtime ID of the air block in the latest version of the game.
	latestAirRID, _ = latestmappings.StateToRuntimeID("minecraft:air", nil)
	// legacyAirRID is the runtime ID of the air block in the v1.12.0 version.
	legacyAirRID = legacymappings.StateToRuntimeID("minecraft:air", nil)
)

// downgradeItem downgrades the input item stack to a legacy item stack. It returns a boolean indicating if the item was
// downgraded successfully.
func downgradeItem(input protocol.ItemStack) legacyprotocol.ItemStack {
	name, _ := latestmappings.ItemRuntimeIDToName(input.NetworkID)
	networkID, _ := legacymappings.ItemIDByName(name)
	return legacyprotocol.ItemStack{
		ItemType: legacyprotocol.ItemType{
			NetworkID:     int32(networkID),
			MetadataValue: int16(input.MetadataValue),
		},
		Count:         int16(input.Count),
		NBTData:       input.NBTData,
		CanBePlacedOn: input.CanBePlacedOn,
		CanBreak:      input.CanBreak,
	}
}

// upgradeItem upgrades the input item stack to a v1.19.0 item stack. It returns a boolean indicating if the item was
// upgraded successfully.
func upgradeItem(input legacyprotocol.ItemStack) protocol.ItemStack {
	if input.ItemType.NetworkID == 0 {
		return protocol.ItemStack{}
	}
	name, _ := legacymappings.ItemNameByID(int16(input.ItemType.NetworkID))
	networkID, _ := latestmappings.ItemNameToRuntimeID(name)
	return protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     networkID,
			MetadataValue: uint32(input.ItemType.MetadataValue),
		},
		Count:         uint16(input.Count),
		NBTData:       input.NBTData,
		CanBePlacedOn: input.CanBePlacedOn,
		CanBreak:      input.CanBreak,
	}
}

// downgradeBlockRuntimeID downgrades a v1.19.0 block runtime ID to a v1.12.0 block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	name, properties, ok := latestmappings.RuntimeIDToState(input)
	if !ok {
		return legacyAirRID
	}
	return legacymappings.StateToRuntimeID(name, properties)
}

// upgradeBlockRuntimeID upgrades a v1.12.0 block runtime ID to a v1.19.0 block runtime ID.
func upgradeBlockRuntimeID(input uint32) uint32 {
	name, properties, ok := legacymappings.RuntimeIDToState(input)
	if !ok {
		return latestAirRID
	}
	runtimeID, ok := latestmappings.StateToRuntimeID(name, properties)
	if !ok {
		return latestAirRID
	}
	return runtimeID
}

// downgradeChunk downgrades a chunk from the latest version to the v1.12.0 equivalent.
func downgradeChunk(chunk *chunk.Chunk) *legacychunk.Chunk {
	// First downgrade the blocks.
	downgraded := legacychunk.New(legacyAirRID)
	for subInd, sub := range chunk.Sub()[4 : len(chunk.Sub())-4] {
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

						downgradedLayer.SetRuntimeID(x, y, z, downgradeBlockRuntimeID(latestRuntimeID))
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
