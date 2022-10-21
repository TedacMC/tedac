package tedac

import (
	"bytes"
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
	"image/color"
)

// Protocol represents the v1.16.100 Protocol implementation.
type Protocol struct{}

// ID ...
func (Protocol) ID() int32 {
	return 419
}

// Ver ...
func (Protocol) Ver() string {
	return "1.16.100"
}

// Packets ...
func (Protocol) Packets() packet.Pool {
	pool := packet.NewPool()
	pool[packet.IDActorPickRequest] = func() packet.Packet { return &legacypacket.ActorPickRequest{} }
	pool[packet.IDCraftingEvent] = func() packet.Packet { return &legacypacket.CraftingEvent{} }
	pool[packet.IDMapInfoRequest] = func() packet.Packet { return &legacypacket.MapInfoRequest{} }
	pool[packet.IDMobArmourEquipment] = func() packet.Packet { return &legacypacket.MobArmourEquipment{} }
	pool[packet.IDMobEquipment] = func() packet.Packet { return &legacypacket.MobEquipment{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDPlayerSkin] = func() packet.Packet { return &legacypacket.PlayerSkin{} }

	pool[packet.IDInventoryTransaction] = func() packet.Packet { return &legacypacket.InventoryTransaction{} }
	return pool
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	return newCFBEncryption(key[:])
}

// nullBytes contains the word 'null' converted to a byte slice.
var nullBytes = []byte("null\n")

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	fmt.Printf("1.16.100 -> 1.19.30: %T\n", pk)
	switch pk := pk.(type) {
	case *legacypacket.ActorPickRequest:
		return []packet.Packet{
			&packet.ActorPickRequest{
				EntityUniqueID: pk.EntityUniqueID,
				HotBarSlot:     pk.HotBarSlot,
			},
		}
	case *legacypacket.CraftingEvent:
		return []packet.Packet{
			&packet.CraftingEvent{
				WindowID:     pk.WindowID,
				CraftingType: pk.CraftingType,
				RecipeUUID:   pk.RecipeUUID,
				Input: lo.Map(pk.Input, func(stack legacyprotocol.ItemStack, _ int) protocol.ItemInstance {
					return protocol.ItemInstance{Stack: upgradeItem(stack)}
				}),
				Output: lo.Map(pk.Output, func(stack legacyprotocol.ItemStack, _ int) protocol.ItemInstance {
					return protocol.ItemInstance{Stack: upgradeItem(stack)}
				}),
			},
		}
	case *legacypacket.MapInfoRequest:
		return []packet.Packet{
			&packet.MapInfoRequest{
				MapID: pk.MapID,
			},
		}
	case *legacypacket.MobArmourEquipment:
		return []packet.Packet{
			&packet.MobArmourEquipment{
				EntityRuntimeID: pk.EntityRuntimeID,
				Helmet:          protocol.ItemInstance{Stack: upgradeItem(pk.Helmet)},
				Chestplate:      protocol.ItemInstance{Stack: upgradeItem(pk.Chestplate)},
				Leggings:        protocol.ItemInstance{Stack: upgradeItem(pk.Leggings)},
				Boots:           protocol.ItemInstance{Stack: upgradeItem(pk.Boots)},
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
	case *legacypacket.PlayerAction:
		return []packet.Packet{
			&packet.PlayerAction{
				EntityRuntimeID: pk.EntityRuntimeID,
				ActionType:      pk.ActionType,
				BlockPosition:   pk.BlockPosition,
				BlockFace:       pk.BlockFace,
			},
		}
	case *legacypacket.PlayerAuthInput:
		return []packet.Packet{
			&packet.PlayerAuthInput{
				Pitch:            pk.Pitch,
				Yaw:              pk.Yaw,
				Position:         pk.Position,
				MoveVector:       pk.MoveVector,
				HeadYaw:          pk.HeadYaw,
				InputData:        pk.InputData,
				InputMode:        pk.InputMode,
				PlayMode:         pk.PlayMode,
				InteractionModel: packet.InteractionModelCrosshair,
				GazeDirection:    pk.GazeDirection,
				Tick:             pk.Tick,
				Delta:            pk.Delta,
			},
		}
	case *legacypacket.PlayerSkin:
		return []packet.Packet{
			&packet.PlayerSkin{
				UUID:        pk.UUID,
				Skin:        legacyprotocol.LatestSkin(pk.Skin),
				NewSkinName: pk.NewSkinName,
				OldSkinName: pk.OldSkinName,
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
	case *packet.AdventureSettings:
	case *packet.TickSync:
		return nil
	case *packet.PacketViolationWarning:
		fmt.Println(pk)
	}
	if pk.ID() == 37 {
		return nil
	}
	return []packet.Packet{pk}
}

// ConvertFromLatest ...
func (Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	fmt.Printf("1.19.30 -> 1.16.100: %T\n", pk)
	switch pk := pk.(type) {
	case *packet.AddActor:
		return []packet.Packet{
			&legacypacket.AddActor{
				EntityUniqueID:  pk.EntityUniqueID,
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityType:      pk.EntityType,
				Position:        pk.Position,
				Velocity:        pk.Velocity,
				Pitch:           pk.Pitch,
				Yaw:             pk.Yaw,
				HeadYaw:         pk.HeadYaw,
				Attributes: lo.Map(pk.Attributes, func(a protocol.AttributeValue, _ int) legacyprotocol.Attribute {
					return legacyprotocol.Attribute{
						Name:  a.Name,
						Value: a.Value,
						Max:   a.Max,
						Min:   a.Min,
					}
				}),
				EntityMetadata: pk.EntityMetadata,
				EntityLinks:    pk.EntityLinks,
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
				EntityMetadata:         pk.EntityMetadata,
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
				EntityLinks:            pk.EntityLinks,
				DeviceID:               pk.DeviceID,
				BuildPlatform:          pk.BuildPlatform,
			},
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
						Aliases:         c.Aliases,
						Overloads: lo.Map(c.Overloads, func(o protocol.CommandOverload, i int) legacyprotocol.CommandOverload {
							return legacyprotocol.CommandOverload{Parameters: lo.Map(o.Parameters, func(p protocol.CommandParameter, _ int) legacyprotocol.CommandParameter {
								return legacyprotocol.CommandParameter{
									Name:                p.Name,
									Type:                p.Type,
									Optional:            p.Optional,
									CollapseEnumOptions: true,
									Enum:                legacyprotocol.CommandEnum(p.Enum),
									Suffix:              p.Suffix,
								}
							})}
						}),
					}
				}),
			},
		}
	case *packet.ClientBoundMapItemData:
		pixels := make([][]color.RGBA, pk.Height)
		pixels = append(pixels, pk.Pixels)
		return []packet.Packet{
			&legacypacket.ClientBoundMapItemData{
				MapID:          pk.MapID,
				UpdateFlags:    pk.UpdateFlags,
				Dimension:      pk.Dimension,
				LockedMap:      pk.LockedMap,
				Scale:          pk.Scale,
				MapsIncludedIn: pk.MapsIncludedIn,
				TrackedObjects: pk.TrackedObjects,
				Decorations:    pk.Decorations,
				Height:         pk.Height,
				Width:          pk.Width,
				XOffset:        pk.XOffset,
				YOffset:        pk.YOffset,
				Pixels:         pixels,
			},
		}
	case *packet.CraftingData:
		return []packet.Packet{
			&legacypacket.CraftingData{
				Recipes:                      []protocol.Recipe{},
				PotionRecipes:                []protocol.PotionRecipe{},
				PotionContainerChangeRecipes: []protocol.PotionContainerChangeRecipe{},
				ClearRecipes:                 pk.ClearRecipes,
				//TODO: Translate these
				//Recipes:                      pk.Recipes,
				//PotionRecipes:                pk.PotionRecipes,
				//PotionContainerChangeRecipes: pk.PotionContainerChangeRecipes,
				//ClearRecipes:                 pk.ClearRecipes,
			},
		}
	case *packet.CraftingEvent:
		return []packet.Packet{
			&legacypacket.CraftingEvent{
				WindowID:     pk.WindowID,
				CraftingType: pk.CraftingType,
				RecipeUUID:   pk.RecipeUUID,
				Input: lo.Map(pk.Input, func(instance protocol.ItemInstance, _ int) legacyprotocol.ItemStack {
					return downgradeItem(instance.Stack)
				}),
				Output: lo.Map(pk.Output, func(instance protocol.ItemInstance, _ int) legacyprotocol.ItemStack {
					return downgradeItem(instance.Stack)
				}),
			},
		}
	case *packet.CreativeContent:
		return []packet.Packet{
			&legacypacket.CreativeContent{
				Items: lo.Map(pk.Items, func(instance protocol.CreativeItem, _ int) legacyprotocol.CreativeItem {
					return legacyprotocol.CreativeItem{
						//CreativeItemNetworkID: instance.CreativeItemNetworkID,
						Item: downgradeItem(instance.Item),
					}
				}),
			},
		}
	case *packet.Event:
		return []packet.Packet{
			&legacypacket.Event{
				EntityRuntimeID: pk.EntityRuntimeID,
				EventType:       pk.EventType,
				UsePlayerID:     pk.UsePlayerID,
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
	case *packet.HurtArmour:
		return []packet.Packet{
			&legacypacket.HurtArmour{
				Cause:  pk.Cause,
				Damage: pk.Damage,
			},
		}
	case *packet.InventoryContent:
		return []packet.Packet{
			&legacypacket.InventoryContent{
				WindowID: pk.WindowID,
				Content: lo.Map(pk.Content, func(instance protocol.ItemInstance, _ int) legacyprotocol.ItemInstance {
					return legacyprotocol.ItemInstance{
						Stack: downgradeItem(instance.Stack),
					}
				}),
			},
		}
	case *packet.InventorySlot:
		return []packet.Packet{
			&legacypacket.InventorySlot{
				WindowID: pk.WindowID,
				Slot:     pk.Slot,
				NewItem: legacyprotocol.ItemInstance{
					Stack: downgradeItem(pk.NewItem.Stack),
				},
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
	case *packet.LevelEvent:
		if pk.EventType == packet.LevelEventParticlesDestroyBlock || pk.EventType == packet.LevelEventParticlesCrackBlock {
			pk.EventData = int32(downgradeBlockRuntimeID(uint32(pk.EventData)))
		}
	case *packet.MapInfoRequest:
		return []packet.Packet{
			&legacypacket.MapInfoRequest{
				MapID: pk.MapID,
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
	case *packet.NetworkChunkPublisherUpdate:
		return []packet.Packet{
			&legacypacket.NetworkChunkPublisherUpdate{
				Position: pk.Position,
				Radius:   pk.Radius,
			},
		}
	case *packet.NetworkSettings:
		return []packet.Packet{
			&legacypacket.NetworkSettings{
				CompressionThreshold: pk.CompressionThreshold,
			},
		}
	case *packet.PlayerList:
		return []packet.Packet{
			&legacypacket.PlayerList{
				ActionType: pk.ActionType,
				Entries: lo.Map(pk.Entries, func(e protocol.PlayerListEntry, _ int) legacypacket.PlayerListEntry {
					return legacypacket.PlayerListEntry{
						UUID:           e.UUID,
						EntityUniqueID: e.EntityUniqueID,
						Username:       e.Username,
						XUID:           e.XUID,
						PlatformChatID: e.PlatformChatID,
						BuildPlatform:  e.BuildPlatform,
						Skin:           legacyprotocol.LegacySkin(e.Skin),
						Teacher:        e.Teacher,
						Host:           e.Host,
					}
				}),
			},
		}
	case *packet.PlayerSkin:
		return []packet.Packet{
			&legacypacket.PlayerSkin{
				UUID:        pk.UUID,
				Skin:        legacyprotocol.LegacySkin(pk.Skin),
				NewSkinName: pk.NewSkinName,
				OldSkinName: pk.OldSkinName,
			},
		}
	case *packet.ResourcePacksInfo:
		return []packet.Packet{
			&legacypacket.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				BehaviourPacks:      pk.BehaviourPacks,
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
				SpawnBiomeType:                 pk.SpawnBiomeType,
				UserDefinedBiomeName:           pk.UserDefinedBiomeName,
				Dimension:                      pk.Dimension,
				Generator:                      pk.Generator,
				WorldGameMode:                  pk.WorldGameMode,
				Difficulty:                     pk.Difficulty,
				WorldSpawn:                     pk.WorldSpawn,
				AchievementsDisabled:           pk.AchievementsDisabled,
				DayCycleLockTime:               pk.DayCycleLockTime,
				EducationEditionOffer:          pk.EducationEditionOffer,
				EducationFeaturesEnabled:       pk.EducationFeaturesEnabled,
				EducationProductID:             pk.EducationProductID,
				RainLevel:                      pk.RainLevel,
				LightningLevel:                 pk.LightningLevel,
				ConfirmedPlatformLockedContent: pk.ConfirmedPlatformLockedContent,
				MultiPlayerGame:                pk.MultiPlayerGame,
				LANBroadcastEnabled:            pk.LANBroadcastEnabled,
				XBLBroadcastMode:               pk.XBLBroadcastMode,
				PlatformBroadcastMode:          pk.PlatformBroadcastMode,
				CommandsEnabled:                pk.CommandsEnabled,
				TexturePackRequired:            pk.TexturePackRequired,
				GameRules: lo.SliceToMap(pk.GameRules, func(rule protocol.GameRule) (string, any) {
					return rule.Name, rule.Value
				}),
				Experiments:                     pk.Experiments,
				ExperimentsPreviouslyToggled:    pk.ExperimentsPreviouslyToggled,
				BonusChestEnabled:               pk.BonusChestEnabled,
				StartWithMapEnabled:             pk.StartWithMapEnabled,
				PlayerPermissions:               pk.PlayerPermissions,
				ServerChunkTickRadius:           pk.ServerChunkTickRadius,
				HasLockedBehaviourPack:          pk.HasLockedBehaviourPack,
				HasLockedTexturePack:            pk.HasLockedTexturePack,
				FromLockedWorldTemplate:         pk.FromLockedWorldTemplate,
				MSAGamerTagsOnly:                pk.MSAGamerTagsOnly,
				FromWorldTemplate:               pk.FromWorldTemplate,
				WorldTemplateSettingsLocked:     pk.WorldTemplateSettingsLocked,
				OnlySpawnV1Villagers:            pk.OnlySpawnV1Villagers,
				BaseGameVersion:                 pk.BaseGameVersion,
				LimitedWorldWidth:               pk.LimitedWorldWidth,
				LimitedWorldDepth:               pk.LimitedWorldDepth,
				NewNether:                       pk.NewNether,
				ForceExperimentalGameplay:       pk.ForceExperimentalGameplay,
				LevelID:                         pk.LevelID,
				WorldName:                       pk.WorldName,
				TemplateContentIdentity:         pk.TemplateContentIdentity,
				Trial:                           pk.Trial,
				ServerAuthoritativeMovementMode: uint32(pk.PlayerMovementSettings.MovementType),
				Time:                            pk.Time,
				EnchantmentSeed:                 pk.EnchantmentSeed,
				MultiPlayerCorrelationID:        pk.MultiPlayerCorrelationID,
				Blocks:                          legacymappings.Blocks(),
				Items:                           legacymappings.Items(),
				ServerAuthoritativeInventory:    pk.ServerAuthoritativeInventory,
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
				Tick: pk.Tick,
			},
		}
	case *packet.UpdateBlock:
		pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
	case *packet.UpdateBlockSynced:
		pk.NewBlockRuntimeID = downgradeBlockRuntimeID(pk.NewBlockRuntimeID)
	case *packet.UpdateAdventureSettings:
		return nil
	}
	return []packet.Packet{pk}
}

var (
	// latestAirRID is the runtime ID of the air block in the latest version of the game.
	latestAirRID, _ = latestmappings.StateToRuntimeID("minecraft:air", nil)
	// legacyAirRID is the runtime ID of the air block in the v1.16.100 version.
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

// downgradeBlockRuntimeID downgrades a v1.19.0 block runtime ID to a v1.16.100 block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	name, properties, ok := latestmappings.RuntimeIDToState(input)
	if !ok {
		return legacyAirRID
	}
	return legacymappings.StateToRuntimeID(name, properties)
}

// upgradeBlockRuntimeID upgrades a v1.16.100 block runtime ID to a v1.19.0 block runtime ID.
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

// downgradeChunk downgrades a chunk from the latest version to the v1.16.100 equivalent.
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
