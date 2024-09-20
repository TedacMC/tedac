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
func (Protocol) Packets(bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDCommandRequest] = func() packet.Packet { return &legacypacket.CommandRequest{} }
	pool[packet.IDContainerClose] = func() packet.Packet { return &legacypacket.ContainerClose{} }
	pool[packet.IDInventoryTransaction] = func() packet.Packet { return &legacypacket.InventoryTransaction{} }
	pool[packet.IDMobEquipment] = func() packet.Packet { return &legacypacket.MobEquipment{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	pool[packet.IDMovePlayer] = func() packet.Packet { return &legacypacket.MovePlayer{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	pool[packet.IDDisconnect] = func() packet.Packet { return &legacypacket.Disconnect{} }
	pool[packet.IDRequestChunkRadius] = func() packet.Packet { return &legacypacket.RequestChunkRadius{} }
	pool[packet.IDText] = func() packet.Packet { return &legacypacket.Text{} }
	pool[packet.IDStopSound] = func() packet.Packet { return &legacypacket.StopSound{} }
	pool[packet.IDSetTitle] = func() packet.Packet { return &legacypacket.SetTitle{} }
	return pool
}

// NewReader ...
func (Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return protocol.NewReader(r, shieldID, enableLimits)
}

// NewWriter ...
func (Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return protocol.NewWriter(w, shieldID)
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return newCFBEncryption(key[:])
}

// nullBytes contains the word 'null' converted to a byte slice.
var nullBytes = []byte("null\n")

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	// fmt.Printf("1.12 -> Latest: %T\n", pk)
	switch pk := pk.(type) {
	case *legacypacket.SetTitle:
		return []packet.Packet{
			&packet.SetTitle{
				ActionType:      pk.ActionType,
				Text:            pk.Text,
				FadeInDuration:  pk.FadeInDuration,
				RemainDuration:  pk.RemainDuration,
				FadeOutDuration: pk.FadeOutDuration,
			},
		}
	case *legacypacket.StopSound:
		return []packet.Packet{
			&packet.StopSound{
				SoundName: pk.SoundName,
				StopAll:   pk.StopAll,
			},
		}
	case *legacypacket.Disconnect:
		return []packet.Packet{
			&packet.Disconnect{
				HideDisconnectionScreen: pk.HideDisconnectionScreen,
				Message:                 pk.Message,
			},
		}
	case *legacypacket.RequestChunkRadius:
		return []packet.Packet{
			&packet.RequestChunkRadius{
				ChunkRadius: pk.ChunkRadius,
			},
		}
	case *legacypacket.Text:
		return []packet.Packet{
			&packet.Text{
				TextType:         pk.TextType,
				NeedsTranslation: pk.NeedsTranslation,
				SourceName:       pk.SourceName,
				Message:          pk.Message,
				Parameters:       pk.Parameters,
				XUID:             pk.XUID,
				PlatformChatID:   pk.PlatformChatID,
			},
		}
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
		var response protocol.Optional[[]byte]
		var cancelReason protocol.Optional[uint8]
		if !bytes.Equal(pk.ResponseData, nullBytes) {
			// The response data is not null, so it is a valid response.
			response = protocol.Option(pk.ResponseData)
		} else {
			// We can always default to the user closed reason if the response data doesn't exist.
			cancelReason = protocol.Option[uint8](packet.ModalFormCancelReasonUserClosed)
		}
		return []packet.Packet{
			&packet.ModalFormResponse{
				FormID:       pk.FormID,
				ResponseData: response,
				CancelReason: cancelReason,
			},
		}
	case *legacypacket.MobEquipment:
		return []packet.Packet{
			&packet.MobEquipment{
				EntityRuntimeID: pk.EntityRuntimeID,
				NewItem:         protocol.ItemInstance{Stack: upgradeItem(pk.NewItem)},
				InventorySlot:   pk.InventorySlot,
				HotBarSlot:      pk.HotBarSlot,
				WindowID:        pk.WindowID,
			},
		}
	case *legacypacket.ContainerClose:
		return []packet.Packet{
			&packet.ContainerClose{
				WindowID:   pk.WindowID,
				ServerSide: false,
			},
		}
	case *legacypacket.CommandRequest:
		return []packet.Packet{
			&packet.CommandRequest{
				CommandLine:   pk.CommandLine,
				CommandOrigin: pk.CommandOrigin,
				Internal:      pk.Internal,
			},
		}
	case *packet.AdventureSettings:
		// TODO: Send request ability instead?
		return nil
	}

	if pk.ID() == 37 { // TODO: This is so fucking ugly why just why
		return nil
	}
	return []packet.Packet{pk}
}

// ConvertFromLatest ...
func (Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	// fmt.Printf("Latest -> 1.12: %T\n", pk)
	switch pk := pk.(type) {
	case *packet.RequestNetworkSettings:
		return []packet.Packet{
			&packet.RequestNetworkSettings{ClientProtocol: protocol.CurrentProtocol},
		}
	case *packet.Transfer:
		return []packet.Packet{
			&legacypacket.Transfer{
				Address: pk.Address,
				Port:    pk.Port,
			},
		}
	case *packet.SetTitle:
		return []packet.Packet{
			&legacypacket.SetTitle{
				ActionType:      pk.ActionType,
				Text:            pk.Text,
				FadeInDuration:  pk.FadeInDuration,
				RemainDuration:  pk.RemainDuration,
				FadeOutDuration: pk.FadeOutDuration,
			},
		}
	case *packet.StopSound:
		return []packet.Packet{
			&legacypacket.StopSound{
				SoundName: pk.SoundName,
				StopAll:   pk.StopAll,
			},
		}
	case *packet.Text:
		return []packet.Packet{
			&legacypacket.Text{
				TextType:         pk.TextType,
				NeedsTranslation: pk.NeedsTranslation,
				SourceName:       pk.SourceName,
				Message:          pk.Message,
				Parameters:       pk.Parameters,
				XUID:             pk.XUID,
				PlatformChatID:   pk.PlatformChatID,
			},
		}
	case *packet.Disconnect:
		return []packet.Packet{
			&legacypacket.Disconnect{
				HideDisconnectionScreen: pk.HideDisconnectionScreen,
				Message:                 pk.Message,
			},
		}
	case *packet.RequestChunkRadius:
		return []packet.Packet{
			&legacypacket.RequestChunkRadius{
				ChunkRadius: pk.ChunkRadius,
			},
		}
	case *packet.StartGame:
		// TODO: Adjust our mappings to account for any possible custom blocks.
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
				EducationFeaturesEnabled:       pk.EducationFeaturesEnabled,
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
				PlayerPermissions:              pk.PlayerPermissions,
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
				PremiumWorldTemplateID:         pk.TemplateContentIdentity,
				Trial:                          pk.Trial,
				Time:                           pk.Time,
				EnchantmentSeed:                pk.EnchantmentSeed,
				MultiPlayerCorrelationID:       pk.MultiPlayerCorrelationID,
				Blocks:                         legacymappings.Blocks(),
				Items:                          legacymappings.Items(),
				GameRules: lo.SliceToMap(pk.GameRules, func(rule protocol.GameRule) (string, any) {
					return rule.Name, rule.Value
				}),
			},
		}
	case *packet.LevelChunk:
		// TODO: Support other sub-chunk request modes.
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
	case *packet.GameRulesChanged:
		return []packet.Packet{
			&legacypacket.GameRulesChanged{
				GameRules: lo.SliceToMap(pk.GameRules, func(rule protocol.GameRule) (string, any) {
					return rule.Name, rule.Value
				}),
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
		return []packet.Packet{
			&legacypacket.AddActor{
				EntityMetadata:  legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityType:      pk.EntityType,
				EntityUniqueID:  pk.EntityUniqueID,
				HeadYaw:         pk.HeadYaw,
				Pitch:           pk.Pitch,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Yaw:             pk.Yaw,
				Attributes: lo.Map(pk.Attributes, func(a protocol.AttributeValue, _ int) legacyprotocol.Attribute {
					return legacyprotocol.Attribute{
						Name:  a.Name,
						Value: a.Value,
						Max:   a.Max,
						Min:   a.Min,
					}
				}),
				EntityLinks: lo.Map(pk.EntityLinks, func(l protocol.EntityLink, _ int) legacyprotocol.EntityLink {
					return legacyprotocol.EntityLink{
						Type:                 l.Type,
						RiddenEntityUniqueID: l.RiddenEntityUniqueID,
						RiderEntityUniqueID:  l.RiddenEntityUniqueID,
						Immediate:            l.Immediate,
					}
				}),
			},
		}
	case *packet.AddPlayer:
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
				EntityMetadata:         legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				DeviceID:               pk.DeviceID,
				EntityLinks: lo.Map(pk.EntityLinks, func(l protocol.EntityLink, _ int) legacyprotocol.EntityLink {
					return legacyprotocol.EntityLink{
						Type:                 l.Type,
						RiddenEntityUniqueID: l.RiddenEntityUniqueID,
						RiderEntityUniqueID:  l.RiddenEntityUniqueID,
						Immediate:            l.Immediate,
					}
				}),
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
				EntityMetadata:  legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
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
		return []packet.Packet{
			&legacypacket.PlayerList{
				ActionType: pk.ActionType,
				Entries: lo.Map(pk.Entries, func(e protocol.PlayerListEntry, _ int) legacypacket.PlayerListEntry {
					var patch struct {
						Geometry struct {
							Default string
						}
					}
					_ = json.Unmarshal(e.Skin.SkinResourcePatch, &patch)
					return legacypacket.PlayerListEntry{
						UUID:             e.UUID,
						EntityUniqueID:   e.EntityUniqueID,
						Username:         e.Username,
						SkinID:           e.Skin.SkinID,
						SkinData:         e.Skin.SkinData,
						CapeData:         e.Skin.CapeData,
						SkinGeometryName: patch.Geometry.Default,
						SkinGeometry:     e.Skin.SkinGeometry,
						PlatformChatID:   e.PlatformChatID,
						XUID:             e.XUID,
					}
				}),
			},
		}
	case *packet.UpdateAbilities:
		if len(pk.AbilityData.Layers) == 0 || pk.AbilityData.EntityUniqueID != conn.GameData().EntityUniqueID {
			// We need at least one layer.
			return nil
		}

		base, flags, perms := pk.AbilityData.Layers[0].Values, uint32(0), uint32(0)
		if base&protocol.AbilityMayFly != 0 {
			flags |= packet.AdventureFlagAllowFlight
			if base&protocol.AbilityFlying != 0 {
				flags |= packet.AdventureFlagFlying
			}
		}
		if base&protocol.AbilityNoClip != 0 {
			flags |= packet.AdventureFlagNoClip
		}

		if base&protocol.AbilityBuild != 0 && base&protocol.AbilityMine != 0 {
			flags |= packet.AdventureFlagWorldBuilder
		} else {
			flags |= packet.AdventureFlagWorldImmutable
		}
		if base&protocol.AbilityBuild != 0 {
			perms |= packet.ActionPermissionBuild
		}
		if base&protocol.AbilityMine != 0 {
			perms |= packet.ActionPermissionMine
		}

		if base&protocol.AbilityDoorsAndSwitches != 0 {
			perms |= packet.ActionPermissionDoorsAndSwitches
		}
		if base&protocol.AbilityOpenContainers != 0 {
			perms |= packet.ActionPermissionOpenContainers
		}
		if base&protocol.AbilityAttackPlayers != 0 {
			perms |= packet.ActionPermissionAttackPlayers
		}
		if base&protocol.AbilityAttackMobs != 0 {
			perms |= packet.ActionPermissionAttackMobs
		}
		return []packet.Packet{
			&packet.AdventureSettings{
				Flags:                  flags,
				ActionPermissions:      perms,
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
				EntityMetadata:  legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
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
	case *packet.AvailableCommands:
		return []packet.Packet{
			&legacypacket.AvailableCommands{
				Commands: lo.Map(pk.Commands, func(c protocol.Command, _ int) legacyprotocol.Command {
					return legacyprotocol.Command{
						Name:            c.Name,
						Description:     c.Description,
						Flags:           byte(c.Flags),
						PermissionLevel: c.PermissionLevel,
						//Aliases:         c.Aliases,
						Overloads: lo.Map(c.Overloads, func(o protocol.CommandOverload, i int) legacyprotocol.CommandOverload {
							return legacyprotocol.CommandOverload{Parameters: lo.Map(o.Parameters, func(p protocol.CommandParameter, _ int) legacyprotocol.CommandParameter {
								return legacyprotocol.CommandParameter{
									Name:                p.Name,
									Type:                p.Type,
									Optional:            p.Optional,
									CollapseEnumOptions: true,
									//Enum:                legacyprotocol.CommandEnum(p.Enum),
									//Suffix:              p.Suffix,
								}
							})}
						}),
					}
				}),
			},
		}
	case *packet.CreativeContent:
		return []packet.Packet{
			&legacypacket.InventoryContent{
				WindowID: 121,
				Content: lo.Map(pk.Items, func(instance protocol.CreativeItem, _ int) legacyprotocol.ItemStack {
					return downgradeItem(instance.Item)
				}),
			},
		}
	case *packet.LevelSoundEvent:
		if pk.SoundType == 113 || pk.SoundType == 145 || pk.SoundType == 151 || pk.SoundType <= 198 && pk.SoundType >= 195 || pk.SoundType == 222 || pk.SoundType == 227 {
			return nil
		}
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
	case *packet.Animate:
		if pk.ActionType > 4 { // TODO: This is also pretty fucking ugly
			return []packet.Packet{}
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

// upgradeItem upgrades the input item stack to the latest item stack. It returns a boolean indicating if the item was
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

// downgradeBlockRuntimeID downgrades the latest block runtime ID to a v1.12.0 block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	name, properties, ok := latestmappings.RuntimeIDToState(input)
	if !ok {
		return legacyAirRID
	}
	return legacymappings.StateToRuntimeID(name, properties)
}

// upgradeBlockRuntimeID upgrades a v1.12.0 block runtime ID to the latest block runtime ID.
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
