package legacyprotocol

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	containerArmour         = 6
	containerChest          = 7
	containerBeacon         = 8
	containerFullInventory  = 12
	containerCraftingGrid   = 13
	containerHotbar         = 27
	containerInventory      = 28
	containerOffHand        = 33
	containerCursor         = 58
	containerCreativeOutput = 59
)

// UpgradeContainerID upgrades a container ID from legacy version to latest version.
func UpgradeContainerID(input byte) byte {
	switch input {
	case containerArmour:
		return protocol.ContainerArmor
	case containerChest:
		return protocol.ContainerLevelEntity
	case containerFullInventory:
		return protocol.ContainerCombinedHotBarAndInventory
	case containerBeacon:
		return protocol.ContainerBeaconPayment
	case containerCraftingGrid:
		return protocol.ContainerCraftingInput
	case containerHotbar:
		return protocol.ContainerHotBar
	case containerInventory:
		return protocol.ContainerInventory
	case containerOffHand:
		return protocol.ContainerOffhand
	case containerCursor:
		return protocol.ContainerCursor
	case containerCreativeOutput:
		return protocol.ContainerCreatedOutput
	}
	return input
}

// DowngradeContainerID downgrade a container ID from latest version to legacy version.
func DowngradeContainerID(input byte) byte {
	switch input {
	case protocol.ContainerArmor:
		return containerArmour
	case protocol.ContainerCombinedHotBarAndInventory:
		return containerFullInventory
	case protocol.ContainerLevelEntity:
		return containerChest
	case protocol.ContainerBeaconPayment:
		return containerBeacon
	case protocol.ContainerCraftingInput:
		return containerCraftingGrid
	case protocol.ContainerHotBar:
		return containerHotbar
	case protocol.ContainerInventory:
		return containerInventory
	case protocol.ContainerOffhand:
		return containerOffHand
	case protocol.ContainerCursor:
		return containerCursor
	case protocol.ContainerCreatedOutput:
		return containerCreativeOutput
	}
	return input
}
