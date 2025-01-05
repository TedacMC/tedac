package tedac

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/samber/lo"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
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
func (Protocol) Packets(bool) packet.Pool {
	pool := packet.NewClientPool()
	for k, v := range packet.NewServerPool() {
		pool[k] = v
	}
	pool[packet.IDActorPickRequest] = func() packet.Packet { return &legacypacket.ActorPickRequest{} }
	pool[packet.IDCommandRequest] = func() packet.Packet { return &legacypacket.CommandRequest{} }
	pool[packet.IDContainerClose] = func() packet.Packet { return &legacypacket.ContainerClose{} }
	// pool[packet.IDCraftingEvent] = func() packet.Packet { return &legacypacket.CraftingEvent{} }
	pool[packet.IDInventoryTransaction] = func() packet.Packet { return &legacypacket.InventoryTransaction{} }
	pool[packet.IDItemStackRequest] = func() packet.Packet { return &legacypacket.ItemStackRequest{} }
	pool[packet.IDItemStackResponse] = func() packet.Packet { return &legacypacket.ItemStackResponse{} }
	pool[packet.IDMapInfoRequest] = func() packet.Packet { return &legacypacket.MapInfoRequest{} }
	pool[packet.IDMobArmourEquipment] = func() packet.Packet { return &legacypacket.MobArmourEquipment{} }
	pool[packet.IDMobEffect] = func() packet.Packet { return &legacypacket.MobEffect{} }
	pool[packet.IDMobEquipment] = func() packet.Packet { return &legacypacket.MobEquipment{} }
	pool[packet.IDModalFormResponse] = func() packet.Packet { return &legacypacket.ModalFormResponse{} }
	pool[packet.IDNPCRequest] = func() packet.Packet { return &legacypacket.NPCRequest{} }
	pool[packet.IDPlayerAction] = func() packet.Packet { return &legacypacket.PlayerAction{} }
	pool[packet.IDPlayerAuthInput] = func() packet.Packet { return &legacypacket.PlayerAuthInput{} }
	pool[packet.IDPlayerSkin] = func() packet.Packet { return &legacypacket.PlayerSkin{} }
	pool[packet.IDResourcePacksInfo] = func() packet.Packet { return &legacypacket.ResourcePacksInfo{} }
	pool[packet.IDResourcePackStack] = func() packet.Packet { return &legacypacket.ResourcePackStack{} }
	pool[packet.IDRequestChunkRadius] = func() packet.Packet { return &legacypacket.RequestChunkRadius{} }
	pool[packet.IDSetActorData] = func() packet.Packet { return &legacypacket.SetActorData{} }
	pool[packet.IDSetActorMotion] = func() packet.Packet { return &legacypacket.SetActorMotion{} }
	pool[packet.IDStructureBlockUpdate] = func() packet.Packet { return &legacypacket.StructureBlockUpdate{} }
	pool[packet.IDStructureTemplateDataRequest] = func() packet.Packet { return &legacypacket.StructureTemplateDataRequest{} }
	pool[packet.IDText] = func() packet.Packet { return &legacypacket.Text{} }
	pool[packet.IDEmote] = func() packet.Packet { return &legacypacket.Emote{} }
	pool[packet.IDTransfer] = func() packet.Packet { return &legacypacket.Transfer{} }
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
	// fmt.Printf("1.16.100 -> Latest: %T\n", pk)
	switch pk := pk.(type) {
	case *legacypacket.ActorPickRequest:
		return []packet.Packet{
			&packet.ActorPickRequest{
				EntityUniqueID: pk.EntityUniqueID,
				HotBarSlot:     pk.HotBarSlot,
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
	case *legacypacket.ContainerClose:
		return []packet.Packet{
			&packet.ContainerClose{
				WindowID:   pk.WindowID,
				ServerSide: pk.ServerSide,
			},
		}
	//case *legacypacket.CraftingEvent:
	//	return []packet.Packet{
	//		&packet.CraftingEvent{
	//			WindowID:     pk.WindowID,
	//			CraftingType: pk.CraftingType,
	//			RecipeUUID:   pk.RecipeUUID,
	//			Input: lo.Map(pk.Input, func(stack legacyprotocol.ItemStack, _ int) protocol.ItemInstance {
	//				return protocol.ItemInstance{Stack: upgradeItem(stack)}
	//			}),
	//			Output: lo.Map(pk.Output, func(stack legacyprotocol.ItemStack, _ int) protocol.ItemInstance {
	//				return protocol.ItemInstance{Stack: upgradeItem(stack)}
	//			}),
	//		},
	//	}
	case *legacypacket.Disconnect:
		return []packet.Packet{
			&packet.Disconnect{
				HideDisconnectionScreen: pk.HideDisconnectionScreen,
				Message:                 pk.Message,
			},
		}
	case *legacypacket.Emote:
		return []packet.Packet{
			&packet.Emote{
				EntityRuntimeID: pk.EntityRuntimeID,
				EmoteID:         pk.EmoteID,
				XUID:            pk.XUID,
				PlatformID:      pk.PlatformID,
				Flags:           pk.Flags,
			},
		}
	case *legacypacket.InventoryTransaction:
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
				LegacyRequestID: pk.LegacyRequestID,
				LegacySetItemSlots: lo.Map(pk.LegacySetItemSlots, func(slot protocol.LegacySetItemSlot, _ int) protocol.LegacySetItemSlot {
					return protocol.LegacySetItemSlot{
						ContainerID: legacyprotocol.UpgradeContainerID(slot.ContainerID),
						Slots:       slot.Slots,
					}
				}),
				Actions: lo.Map(pk.Actions, func(action legacyprotocol.InventoryAction, _ int) protocol.InventoryAction {
					return protocol.InventoryAction{
						SourceType:    action.SourceType,
						WindowID:      action.WindowID,
						SourceFlags:   action.SourceFlags,
						InventorySlot: action.InventorySlot,
						OldItem:       protocol.ItemInstance{Stack: upgradeItem(action.OldItem)},
						NewItem:       protocol.ItemInstance{Stack: upgradeItem(action.NewItem)},
					}
				}),
				TransactionData: transactionData,
			},
		}
	case *legacypacket.ItemStackRequest:
		return []packet.Packet{
			&packet.ItemStackRequest{
				Requests: lo.Map(pk.Requests, func(request legacyprotocol.ItemStackRequest, _ int) protocol.ItemStackRequest {
					return protocol.ItemStackRequest{
						RequestID: request.RequestID,
						Actions: lo.Map(request.Actions, func(action legacyprotocol.StackRequestAction, _ int) protocol.StackRequestAction {
							switch data := action.(type) {
							case *legacyprotocol.TakeStackRequestAction:
								newAction := &protocol.TakeStackRequestAction{}
								newAction.Count = data.Count
								newAction.Source = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Source.ContainerID)},
									Slot:           data.Source.Slot,
									StackNetworkID: data.Source.StackNetworkID,
								}
								newAction.Destination = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Destination.ContainerID)},
									Slot:           data.Destination.Slot,
									StackNetworkID: data.Destination.StackNetworkID,
								}
								return newAction
							case *legacyprotocol.PlaceStackRequestAction:
								newAction := &protocol.PlaceStackRequestAction{}
								newAction.Count = data.Count
								newAction.Source = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Source.ContainerID)},
									Slot:           data.Source.Slot,
									StackNetworkID: data.Source.StackNetworkID,
								}
								newAction.Destination = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Destination.ContainerID)},
									Slot:           data.Destination.Slot,
									StackNetworkID: data.Destination.StackNetworkID,
								}
								return newAction
							case *legacyprotocol.SwapStackRequestAction:
								newAction := &protocol.SwapStackRequestAction{}
								newAction.Source = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Source.ContainerID)},
									Slot:           data.Source.Slot,
									StackNetworkID: data.Source.StackNetworkID,
								}
								newAction.Destination = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Destination.ContainerID)},
									Slot:           data.Destination.Slot,
									StackNetworkID: data.Destination.StackNetworkID,
								}
								return newAction
							case *legacyprotocol.DropStackRequestAction:
								newAction := &protocol.DropStackRequestAction{}
								newAction.Count = data.Count
								newAction.Source = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Source.ContainerID)},
									Slot:           data.Source.Slot,
									StackNetworkID: data.Source.StackNetworkID,
								}
								newAction.Randomly = data.Randomly
								return newAction
							case *legacyprotocol.DestroyStackRequestAction:
								newAction := &protocol.DestroyStackRequestAction{}
								newAction.Count = data.Count
								newAction.Source = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Source.ContainerID)},
									Slot:           data.Source.Slot,
									StackNetworkID: data.Source.StackNetworkID,
								}
								return newAction
							case *legacyprotocol.ConsumeStackRequestAction:
								newAction := &protocol.ConsumeStackRequestAction{}
								newAction.Count = data.Count
								newAction.Source = protocol.StackRequestSlotInfo{
									Container:      protocol.FullContainerName{ContainerID: legacyprotocol.UpgradeContainerID(data.Source.ContainerID)},
									Slot:           data.Source.Slot,
									StackNetworkID: data.Source.StackNetworkID,
								}
								return newAction
							case *legacyprotocol.CreateStackRequestAction:
								return &protocol.CreateStackRequestAction{
									ResultsSlot: data.ResultsSlot,
								}
							case *legacyprotocol.LabTableCombineStackRequestAction:
								return &protocol.LabTableCombineStackRequestAction{}
							case *legacyprotocol.BeaconPaymentStackRequestAction:
								return &protocol.BeaconPaymentStackRequestAction{
									PrimaryEffect:   data.PrimaryEffect,
									SecondaryEffect: data.SecondaryEffect,
								}
							case *legacyprotocol.CraftRecipeStackRequestAction:
								return &protocol.CraftRecipeStackRequestAction{
									RecipeNetworkID: data.RecipeNetworkID,
								}
							case *legacyprotocol.AutoCraftRecipeStackRequestAction:
								return &protocol.AutoCraftRecipeStackRequestAction{
									RecipeNetworkID: data.RecipeNetworkID,
								}
							case *legacyprotocol.CraftCreativeStackRequestAction:
								return &protocol.CraftCreativeStackRequestAction{
									CreativeItemNetworkID: data.CreativeItemNetworkID,
								}
							case *legacyprotocol.CraftNonImplementedStackRequestAction:
								return &protocol.CraftNonImplementedStackRequestAction{}
							case *legacyprotocol.CraftResultsDeprecatedStackRequestAction:
								return &protocol.CraftResultsDeprecatedStackRequestAction{
									ResultItems: lo.Map(data.ResultItems, func(stack legacyprotocol.ItemStack, _ int) protocol.ItemStack {
										return upgradeItem(stack)
									}),
								}
							}
							return nil
						}),
					}
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
	case *legacypacket.MobEffect:
		return []packet.Packet{
			&packet.MobEffect{
				EntityRuntimeID: pk.EntityRuntimeID,
				Operation:       pk.Operation,
				EffectType:      pk.EffectType,
				Amplifier:       pk.Amplifier,
				Particles:       pk.Particles,
				Duration:        pk.Duration,
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
	case *legacypacket.NPCRequest:
		return []packet.Packet{
			&packet.NPCRequest{
				EntityRuntimeID: pk.EntityRuntimeID,
				RequestType:     pk.RequestType,
				CommandString:   pk.CommandString,
				ActionType:      pk.ActionType,
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
				InputData:        legacyprotocol.BitSet(pk.InputData, packet.PlayerAuthInputBitsetSize),
				InputMode:        pk.InputMode,
				PlayMode:         pk.PlayMode,
				InteractionModel: packet.InteractionModelCrosshair,
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
	case *legacypacket.ResourcePackStack:
		return []packet.Packet{
			&packet.ResourcePackStack{
				TexturePackRequired:          pk.TexturePackRequired,
				BehaviourPacks:               pk.BehaviourPacks,
				TexturePacks:                 pk.TexturePacks,
				BaseGameVersion:              pk.BaseGameVersion,
				Experiments:                  pk.Experiments,
				ExperimentsPreviouslyToggled: pk.PreviouslyHadExperimentsToggled,
			},
		}
	case *legacypacket.RequestChunkRadius:
		return []packet.Packet{
			&packet.RequestChunkRadius{
				ChunkRadius: pk.ChunkRadius,
			},
		}
	case *legacypacket.SetActorData:
		return []packet.Packet{
			&packet.SetActorData{
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityMetadata:  legacyprotocol.UpgradeEntityMetadata(pk.EntityMetadata),
				Tick:            pk.Tick,
			},
		}
	case *legacypacket.SetActorMotion:
		return []packet.Packet{
			&packet.SetActorMotion{
				EntityRuntimeID: pk.EntityRuntimeID,
				Velocity:        pk.Velocity,
			},
		}
	case *legacypacket.StructureBlockUpdate:
		return []packet.Packet{
			&packet.StructureBlockUpdate{
				Position:           pk.Position,
				StructureName:      pk.StructureName,
				DataField:          pk.DataField,
				IncludePlayers:     pk.IncludePlayers,
				ShowBoundingBox:    pk.ShowBoundingBox,
				StructureBlockType: pk.StructureBlockType,
				Settings: protocol.StructureSettings{
					PaletteName:               pk.Settings.PaletteName,
					IgnoreEntities:            pk.Settings.IgnoreEntities,
					IgnoreBlocks:              pk.Settings.IgnoreBlocks,
					Size:                      pk.Settings.Size,
					Offset:                    pk.Settings.Offset,
					LastEditingPlayerUniqueID: pk.Settings.LastEditingPlayerUniqueID,
					Rotation:                  pk.Settings.Rotation,
					Mirror:                    pk.Settings.Mirror,
					Integrity:                 pk.Settings.Integrity,
					Seed:                      pk.Settings.Seed,
					Pivot:                     pk.Settings.Pivot,
				},
				RedstoneSaveMode: pk.RedstoneSaveMode,
				ShouldTrigger:    pk.ShouldTrigger,
			},
		}
	case *legacypacket.StructureTemplateDataRequest:
		return []packet.Packet{
			&packet.StructureTemplateDataRequest{
				StructureName: pk.StructureName,
				Position:      pk.Position,
				Settings: protocol.StructureSettings{
					PaletteName:               pk.Settings.PaletteName,
					IgnoreEntities:            pk.Settings.IgnoreEntities,
					IgnoreBlocks:              pk.Settings.IgnoreBlocks,
					Size:                      pk.Settings.Size,
					Offset:                    pk.Settings.Offset,
					LastEditingPlayerUniqueID: pk.Settings.LastEditingPlayerUniqueID,
					Rotation:                  pk.Settings.Rotation,
					Mirror:                    pk.Settings.Mirror,
					Integrity:                 pk.Settings.Integrity,
					Seed:                      pk.Settings.Seed,
					Pivot:                     pk.Settings.Pivot,
				},
				RequestType: pk.RequestType,
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
	case *legacypacket.Transfer:
		return []packet.Packet{
			&packet.Transfer{
				Address: pk.Address,
				Port:    pk.Port,
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
	// fmt.Printf("Latest -> 1.16.100: %T\n", pk)
	switch pk := pk.(type) {
	case *packet.ActorEvent:
		if pk.EventType > packet.ActorEventFinishedChargingItem {
			return nil
		}
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
				EntityMetadata: legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
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
				EntityMetadata:  legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
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
				EntityMetadata:         legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
				CommandPermissionLevel: uint32(pk.AbilityData.CommandPermissions),
				PermissionLevel:        uint32(pk.AbilityData.PlayerPermissions),
				PlayerUniqueID:         pk.AbilityData.EntityUniqueID,
				EntityLinks:            pk.EntityLinks,
				DeviceID:               pk.DeviceID,
				BuildPlatform:          pk.BuildPlatform,
			},
		}
	case *packet.AnimateEntity:
		return []packet.Packet{
			&legacypacket.AnimateEntity{
				Animation:        pk.Animation,
				NextState:        pk.NextState,
				StopCondition:    pk.StopCondition,
				Controller:       pk.Controller,
				BlendOutTime:     pk.BlendOutTime,
				EntityRuntimeIDs: pk.EntityRuntimeIDs,
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
						//Aliases:         c.Aliases,
						Overloads: lo.Map(c.Overloads, func(o protocol.CommandOverload, i int) legacyprotocol.CommandOverload {
							return legacyprotocol.CommandOverload{Parameters: lo.Map(o.Parameters, func(p protocol.CommandParameter, _ int) legacyprotocol.CommandParameter {
								return legacyprotocol.CommandParameter{
									Name:                p.Name,
									Type:                legacyprotocol.DowngradeParamType(p.Type),
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
	case *packet.CameraShake:
		return []packet.Packet{
			&legacypacket.CameraShake{
				Intensity: pk.Intensity,
				Duration:  pk.Duration,
				Type:      pk.Type,
			},
		}
	case *packet.ClientBoundMapItemData:
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
				Pixels:         [][]color.RGBA{pk.Pixels},
			},
		}
	case *packet.ContainerClose:
		return []packet.Packet{
			&legacypacket.ContainerClose{
				WindowID:   pk.WindowID,
				ServerSide: pk.ServerSide,
			},
		}
	case *packet.CraftingData:
		return []packet.Packet{
			&legacypacket.CraftingData{
				Recipes: lo.Map(pk.Recipes, func(r protocol.Recipe, _ int) legacyprotocol.Recipe {
					switch data := r.(type) {
					case *protocol.ShapelessRecipe:
						recipe := &legacyprotocol.ShapelessRecipe{}
						recipe.RecipeID = data.RecipeID
						recipe.Input = lo.Map(data.Input, func(item protocol.ItemDescriptorCount, _ int) legacyprotocol.RecipeIngredientItem {
							switch d := item.Descriptor.(type) {
							case *protocol.DefaultItemDescriptor:
								return legacyprotocol.RecipeIngredientItem{
									NetworkID:     int32(d.NetworkID),
									MetadataValue: int32(d.MetadataValue),
									Count:         item.Count,
								}
							}
							return legacyprotocol.RecipeIngredientItem{}
						})
						recipe.Output = lo.Map(data.Output, func(item protocol.ItemStack, _ int) legacyprotocol.ItemStack {
							return downgradeItem(item)
						})
						recipe.UUID = data.UUID
						recipe.Block = data.Block
						recipe.Priority = data.Priority
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					case *protocol.ShulkerBoxRecipe:
						recipe := &legacyprotocol.ShulkerBoxRecipe{}
						recipe.RecipeID = data.RecipeID
						recipe.Input = lo.Map(data.Input, func(item protocol.ItemDescriptorCount, _ int) legacyprotocol.RecipeIngredientItem {
							switch d := item.Descriptor.(type) {
							case *protocol.DefaultItemDescriptor:
								return legacyprotocol.RecipeIngredientItem{
									NetworkID:     int32(d.NetworkID),
									MetadataValue: int32(d.MetadataValue),
									Count:         item.Count,
								}
							}
							return legacyprotocol.RecipeIngredientItem{}
						})
						recipe.Output = lo.Map(data.Output, func(item protocol.ItemStack, _ int) legacyprotocol.ItemStack {
							return downgradeItem(item)
						})
						recipe.UUID = data.UUID
						recipe.Block = data.Block
						recipe.Priority = data.Priority
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					case *protocol.ShapelessChemistryRecipe:
						recipe := &legacyprotocol.ShapelessChemistryRecipe{}
						recipe.RecipeID = data.RecipeID
						recipe.Input = lo.Map(data.Input, func(item protocol.ItemDescriptorCount, _ int) legacyprotocol.RecipeIngredientItem {
							switch d := item.Descriptor.(type) {
							case *protocol.DefaultItemDescriptor:
								return legacyprotocol.RecipeIngredientItem{
									NetworkID:     int32(d.NetworkID),
									MetadataValue: int32(d.MetadataValue),
									Count:         item.Count,
								}
							}
							return legacyprotocol.RecipeIngredientItem{}
						})
						recipe.Output = lo.Map(data.Output, func(item protocol.ItemStack, _ int) legacyprotocol.ItemStack {
							return downgradeItem(item)
						})
						recipe.UUID = data.UUID
						recipe.Block = data.Block
						recipe.Priority = data.Priority
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					case *protocol.ShapedRecipe:
						recipe := &legacyprotocol.ShapedRecipe{}
						recipe.RecipeID = data.RecipeID
						recipe.Width = data.Width
						recipe.Height = data.Height
						recipe.Input = lo.Map(data.Input, func(item protocol.ItemDescriptorCount, _ int) legacyprotocol.RecipeIngredientItem {
							switch d := item.Descriptor.(type) {
							case *protocol.DefaultItemDescriptor:
								return legacyprotocol.RecipeIngredientItem{
									NetworkID:     int32(d.NetworkID),
									MetadataValue: int32(d.MetadataValue),
									Count:         item.Count,
								}
							}
							return legacyprotocol.RecipeIngredientItem{}
						})
						recipe.Output = lo.Map(data.Output, func(item protocol.ItemStack, _ int) legacyprotocol.ItemStack {
							return downgradeItem(item)
						})
						recipe.UUID = data.UUID
						recipe.Block = data.Block
						recipe.Priority = data.Priority
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					case *protocol.ShapedChemistryRecipe:
						recipe := &legacyprotocol.ShapedChemistryRecipe{}
						recipe.RecipeID = data.RecipeID
						recipe.Width = data.Width
						recipe.Height = data.Height
						recipe.Input = lo.Map(data.Input, func(item protocol.ItemDescriptorCount, _ int) legacyprotocol.RecipeIngredientItem {
							switch d := item.Descriptor.(type) {
							case *protocol.DefaultItemDescriptor:
								return legacyprotocol.RecipeIngredientItem{
									NetworkID:     int32(d.NetworkID),
									MetadataValue: int32(d.MetadataValue),
									Count:         item.Count,
								}
							}
							return legacyprotocol.RecipeIngredientItem{}
						})
						recipe.Output = lo.Map(data.Output, func(item protocol.ItemStack, _ int) legacyprotocol.ItemStack {
							return downgradeItem(item)
						})
						recipe.UUID = data.UUID
						recipe.Block = data.Block
						recipe.Priority = data.Priority
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					case *protocol.FurnaceRecipe:
						recipe := &legacyprotocol.FurnaceRecipe{}
						recipe.InputType = legacyprotocol.ItemType{
							NetworkID:     data.InputType.NetworkID,
							MetadataValue: int16(data.InputType.MetadataValue),
						}
						recipe.Output = downgradeItem(data.Output)
						recipe.Block = data.Block
						return recipe
					case *protocol.FurnaceDataRecipe:
						recipe := &legacyprotocol.FurnaceDataRecipe{}
						recipe.InputType = legacyprotocol.ItemType{
							NetworkID:     data.InputType.NetworkID,
							MetadataValue: int16(data.InputType.MetadataValue),
						}
						recipe.Output = downgradeItem(data.Output)
						recipe.Block = data.Block
						return recipe
					case *protocol.MultiRecipe:
						recipe := &legacyprotocol.MultiRecipe{}
						recipe.UUID = data.UUID
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					case *protocol.SmithingTransformRecipe:
						recipe := &legacyprotocol.ShapelessRecipe{}
						recipe.RecipeID = data.RecipeID
						recipe.Input = lo.Map(append([]protocol.ItemDescriptorCount{}, data.Base, data.Addition), func(item protocol.ItemDescriptorCount, _ int) legacyprotocol.RecipeIngredientItem {
							switch d := item.Descriptor.(type) {
							case *protocol.DefaultItemDescriptor:
								return legacyprotocol.RecipeIngredientItem{
									NetworkID:     int32(d.NetworkID),
									MetadataValue: int32(d.MetadataValue),
									Count:         item.Count,
								}
							}
							return legacyprotocol.RecipeIngredientItem{}
						})
						recipe.Output = append([]legacyprotocol.ItemStack{}, downgradeItem(data.Result))
						recipe.Block = data.Block
						recipe.RecipeNetworkID = data.RecipeNetworkID
						return recipe
					}
					return nil
				}),
				PotionRecipes:                pk.PotionRecipes,
				PotionContainerChangeRecipes: pk.PotionContainerChangeRecipes,
				ClearRecipes:                 pk.ClearRecipes,
			},
		}
	case *packet.CreativeContent:
		return []packet.Packet{
			&legacypacket.CreativeContent{
				Items: lo.Map(pk.Items, func(instance protocol.CreativeItem, _ int) legacyprotocol.CreativeItem {
					return legacyprotocol.CreativeItem{
						CreativeItemNetworkID: instance.CreativeItemNetworkID,
						Item:                  downgradeItem(instance.Item),
					}
				}),
			},
		}
	case *packet.Disconnect:
		return []packet.Packet{
			&legacypacket.Disconnect{
				HideDisconnectionScreen: pk.HideDisconnectionScreen,
				Message:                 pk.Message,
			},
		}
	case *packet.EducationSettings:
		return []packet.Packet{
			&legacypacket.EducationSettings{
				CodeBuilderDefaultURI: pk.CodeBuilderDefaultURI,
				CodeBuilderTitle:      pk.CodeBuilderTitle,
				CanResizeCodeBuilder:  pk.CanResizeCodeBuilder,
				OverrideURI:           pk.OverrideURI,
				HasQuiz:               pk.HasQuiz,
			},
		}
	case *packet.Emote:
		return []packet.Packet{
			&legacypacket.Emote{
				EntityRuntimeID: pk.EntityRuntimeID,
				EmoteID:         pk.EmoteID,
				XUID:            pk.XUID,
				PlatformID:      pk.PlatformID,
				Flags:           pk.Flags,
			},
		}
	case *packet.Event:
		// TODO: support
		return []packet.Packet{
			&legacypacket.Event{
				EntityRuntimeID: uint64(pk.EntityRuntimeID),
				EventType:       0,
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
	case *packet.InventoryTransaction:
		var transactionData legacyprotocol.InventoryTransactionData
		switch data := pk.TransactionData.(type) {
		case *protocol.NormalTransactionData:
			transactionData = &legacyprotocol.NormalTransactionData{}
		case *protocol.MismatchTransactionData:
			transactionData = &legacyprotocol.MismatchTransactionData{}
		case *protocol.UseItemTransactionData:
			transactionData = &legacyprotocol.UseItemTransactionData{
				ActionType:      data.ActionType,
				BlockPosition:   data.BlockPosition,
				BlockFace:       data.BlockFace,
				HotBarSlot:      data.HotBarSlot,
				HeldItem:        downgradeItem(data.HeldItem.Stack),
				Position:        data.Position,
				ClickedPosition: data.ClickedPosition,
				BlockRuntimeID:  downgradeBlockRuntimeID(data.BlockRuntimeID),
			}
		case *protocol.UseItemOnEntityTransactionData:
			transactionData = &legacyprotocol.UseItemOnEntityTransactionData{
				TargetEntityRuntimeID: data.TargetEntityRuntimeID,
				ActionType:            data.ActionType,
				HotBarSlot:            data.HotBarSlot,
				HeldItem:              downgradeItem(data.HeldItem.Stack),
				Position:              data.Position,
				ClickedPosition:       data.ClickedPosition,
			}
		case *protocol.ReleaseItemTransactionData:
			transactionData = &legacyprotocol.ReleaseItemTransactionData{
				ActionType:   data.ActionType,
				HotBarSlot:   data.HotBarSlot,
				HeldItem:     downgradeItem(data.HeldItem.Stack),
				HeadPosition: data.HeadPosition,
			}
		}
		return []packet.Packet{
			&legacypacket.InventoryTransaction{
				LegacyRequestID: pk.LegacyRequestID,
				LegacySetItemSlots: lo.Map(pk.LegacySetItemSlots, func(slot protocol.LegacySetItemSlot, _ int) protocol.LegacySetItemSlot {
					return protocol.LegacySetItemSlot{
						ContainerID: legacyprotocol.DowngradeContainerID(slot.ContainerID),
						Slots:       slot.Slots,
					}
				}),
				Actions: lo.Map(pk.Actions, func(action protocol.InventoryAction, _ int) legacyprotocol.InventoryAction {
					return legacyprotocol.InventoryAction{
						SourceType:    action.SourceType,
						WindowID:      action.WindowID,
						SourceFlags:   action.SourceFlags,
						InventorySlot: action.InventorySlot,
						OldItem:       downgradeItem(action.OldItem.Stack),
						NewItem:       downgradeItem(action.NewItem.Stack),
					}
				}),
				TransactionData: transactionData,
			},
		}
	case *packet.ItemStackResponse:
		return []packet.Packet{
			&legacypacket.ItemStackResponse{
				Responses: lo.Map(pk.Responses, func(response protocol.ItemStackResponse, _ int) legacyprotocol.ItemStackResponse {
					return legacyprotocol.ItemStackResponse{
						Status:    response.Status,
						RequestID: response.RequestID,
						ContainerInfo: lo.Map(response.ContainerInfo, func(info protocol.StackResponseContainerInfo, _ int) legacyprotocol.StackResponseContainerInfo {
							return legacyprotocol.StackResponseContainerInfo{
								ContainerID: legacyprotocol.DowngradeContainerID(info.Container.ContainerID),
								SlotInfo: lo.Map(info.SlotInfo, func(slot protocol.StackResponseSlotInfo, _ int) legacyprotocol.StackResponseSlotInfo {
									return legacyprotocol.StackResponseSlotInfo{
										Slot:           slot.Slot,
										HotbarSlot:     slot.HotbarSlot,
										Count:          slot.Count,
										StackNetworkID: slot.StackNetworkID,
									}
								}),
							}
						}),
					}
				}),
			},
		}
	case *packet.LevelChunk:
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
	case *packet.LevelSoundEvent:
		if pk.SoundType == packet.SoundEventPlace || pk.SoundType == packet.SoundEventHit || pk.SoundType == packet.SoundEventItemUseOn {
			pk.ExtraData = int32(downgradeBlockRuntimeID(uint32(pk.ExtraData)))
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
	case *packet.MobEffect:
		return []packet.Packet{
			&legacypacket.MobEffect{
				EntityRuntimeID: pk.EntityRuntimeID,
				Operation:       pk.Operation,
				EffectType:      pk.EffectType,
				Amplifier:       pk.Amplifier,
				Particles:       pk.Particles,
				Duration:        pk.Duration,
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
	case *packet.PhotoTransfer:
		return []packet.Packet{
			&legacypacket.PhotoTransfer{
				PhotoName: pk.PhotoName,
				PhotoData: pk.PhotoData,
				BookID:    pk.BookID,
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
	case *packet.PositionTrackingDBServerBroadcast:
		data, _ := nbt.MarshalEncoding(&pk.Payload, nbt.LittleEndian)
		return []packet.Packet{
			&legacypacket.PositionTrackingDBServerBroadcast{
				BroadcastAction: pk.BroadcastAction,
				TrackingID:      pk.TrackingID,
				SerialisedData:  data,
			},
		}
	case *packet.ResourcePacksInfo:
		return []packet.Packet{
			&legacypacket.ResourcePacksInfo{
				TexturePackRequired: pk.TexturePackRequired,
				HasScripts:          pk.HasScripts,
				TexturePacks: lo.Map(pk.TexturePacks, func(pack protocol.TexturePackInfo, _ int) legacyprotocol.ResourcePackInfo {
					return legacyprotocol.ResourcePackInfo{
						UUID:            pack.UUID.String(),
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
				TexturePackRequired:             pk.TexturePackRequired,
				BehaviourPacks:                  pk.BehaviourPacks,
				TexturePacks:                    pk.TexturePacks,
				BaseGameVersion:                 pk.BaseGameVersion,
				Experiments:                     pk.Experiments,
				PreviouslyHadExperimentsToggled: pk.ExperimentsPreviouslyToggled,
			},
		}
	case *packet.RequestChunkRadius:
		return []packet.Packet{
			&legacypacket.RequestChunkRadius{
				ChunkRadius: pk.ChunkRadius,
			},
		}
	case *packet.SetActorData:
		return []packet.Packet{
			&legacypacket.SetActorData{
				EntityRuntimeID: pk.EntityRuntimeID,
				EntityMetadata:  legacyprotocol.DowngradeEntityMetadata(pk.EntityMetadata),
				Tick:            pk.Tick,
			},
		}
	case *packet.SetActorMotion:
		return []packet.Packet{
			&legacypacket.SetActorMotion{
				EntityRuntimeID: pk.EntityRuntimeID,
				Velocity:        pk.Velocity,
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
	case *packet.SpawnParticleEffect:
		return []packet.Packet{
			&legacypacket.SpawnParticleEffect{
				Dimension:      pk.Dimension,
				EntityUniqueID: pk.EntityUniqueID,
				Position:       pk.Position,
				ParticleName:   pk.ParticleName,
			},
		}
	case *packet.StartGame:
		// TODO: Adjust our mappings to account for any possible custom blocks.
		force, ok := pk.ForceExperimentalGameplay.Value()
		if !ok {
			force = ok
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
				ForceExperimentalGameplay:       force,
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
	case *packet.Transfer:
		return []packet.Packet{
			&legacypacket.Transfer{
				Address: pk.Address,
				Port:    pk.Port,
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

// downgradeBlockRuntimeID downgrades the latest block runtime ID to a v1.16.100 block runtime ID.
func downgradeBlockRuntimeID(input uint32) uint32 {
	name, properties, ok := latestmappings.RuntimeIDToState(input)
	if !ok {
		return legacyAirRID
	}
	return legacymappings.StateToRuntimeID(name, properties)
}

// upgradeBlockRuntimeID upgrades a v1.16.100 block runtime ID to the latest block runtime ID.
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
