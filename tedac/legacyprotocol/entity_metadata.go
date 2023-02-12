package legacyprotocol

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

// DowngradeEntityMetadata downgrades entity metadata from latest version to legacy version.
func DowngradeEntityMetadata(data map[uint32]any) map[uint32]any {
	data = downgradeKey(data)

	var flag1, flag2 int64
	if v, ok := data[protocol.EntityDataKeyFlags]; ok {
		flag1 = v.(int64)
	}
	if v, ok := data[protocol.EntityDataKeyFlagsTwo]; ok {
		flag2 = v.(int64)
	}
	if flag1 == 0 && flag2 == 0 {
		return data
	}

	newFlag1 := flag1 & ^(^0 << (protocol.EntityDataFlagDash - 1))
	lastHalf := flag1 & (^0 << protocol.EntityDataFlagDash)
	lastHalf >>= 1
	lastHalf &= math.MaxInt64

	newFlag1 |= lastHalf

	if flag2 != 0 {
		newFlag1 ^= (flag2 & 1) << 63
		flag2 >>= 1
		flag2 &= math.MaxInt64

		data[protocol.EntityDataKeyFlagsTwo] = flag2
	}

	data[protocol.EntityDataKeyFlags] = newFlag1
	return data
}

// UpgradeEntityMetadata upgrades entity metadata from legacy version to latest version.
func UpgradeEntityMetadata(data map[uint32]any) map[uint32]any {
	data = upgradeKey(data)

	var flag1, flag2 int64
	if v, ok := data[protocol.EntityDataKeyFlags]; ok {
		flag1 = v.(int64)
	}
	if v, ok := data[protocol.EntityDataKeyFlagsTwo]; ok {
		flag2 = v.(int64)
	}

	flag2 <<= 1
	flag2 |= (flag1 >> 63) & 1

	newFlag1 := flag1 & ^(^0 << (protocol.EntityDataFlagDash - 1))
	lastHalf := flag1 & (^0<<protocol.EntityDataFlagDash - 1)
	lastHalf <<= 1
	newFlag1 |= lastHalf

	data[protocol.EntityDataKeyFlagsTwo] = flag2
	data[protocol.EntityDataKeyFlags] = newFlag1
	return data
}

// downgradeKey downgrades the latest key of an entity metadata map to the legacy key.
func downgradeKey(data map[uint32]any) map[uint32]any {
	newData := make(map[uint32]any)
	for key, value := range data {
		switch key {
		case protocol.EntityDataKeyDataRadius:
			key = 60
		case protocol.EntityDataKeyDataWaiting:
			key = 61
		case protocol.EntityDataKeyDataParticle:
			key = 62
		case protocol.EntityDataKeyAttachFace:
			key = 64
		case protocol.EntityDataKeyAttachedPosition:
			key = 66
		case protocol.EntityDataKeyTradeTarget:
			key = 67
		case protocol.EntityDataKeyCommandName:
			key = 70
		case protocol.EntityDataKeyLastCommandOutput:
			key = 71
		case protocol.EntityDataKeyTrackCommandOutput:
			key = 72
		case protocol.EntityDataKeyControllingSeatIndex:
			key = 73
		case protocol.EntityDataKeyStrength:
			key = 74
		case protocol.EntityDataKeyStrengthMax:
			key = 75
		case protocol.EntityDataKeyDataLifetimeTicks:
			key = 77
		case protocol.EntityDataKeyPoseIndex:
			key = 78
		case protocol.EntityDataKeyDataTickOffset:
			key = 79
		case protocol.EntityDataKeyAlwaysShowNameTag:
			key = 80
		case protocol.EntityDataKeyColorTwoIndex:
			key = 81
		case protocol.EntityDataKeyScore:
			key = 83
		case protocol.EntityDataKeyBalloonAnchor:
			key = 84
		case protocol.EntityDataKeyPuffedState:
			key = 85
		case protocol.EntityDataKeyBubbleTime:
			key = 86
		case protocol.EntityDataKeyAgent:
			key = 87
		case protocol.EntityDataKeyEatingCounter:
			key = 90
		case protocol.EntityDataKeyFlagsTwo:
			key = 91
		case protocol.EntityDataKeyDataDuration:
			key = 94
		case protocol.EntityDataKeyDataSpawnTime:
			key = 95
		case protocol.EntityDataKeyDataChangeRate:
			key = 96
		case protocol.EntityDataKeyDataChangeOnPickup:
			key = 97
		case protocol.EntityDataKeyDataPickupCount:
			key = 98
		case protocol.EntityDataKeyInteractText:
			key = 99
		case protocol.EntityDataKeyTradeTier:
			key = 100
		case protocol.EntityDataKeyMaxTradeTier:
			key = 101
		case protocol.EntityDataKeyTradeExperience:
			key = 102
		case protocol.EntityDataKeySkinID:
			key = 104
		case protocol.EntityDataKeyCommandBlockTickDelay:
			key = 105
		case protocol.EntityDataKeyCommandBlockExecuteOnFirstTick:
			key = 106
		case protocol.EntityDataKeyAmbientSoundInterval:
			key = 107
		case protocol.EntityDataKeyAmbientSoundIntervalRange:
			key = 108
		case protocol.EntityDataKeyAmbientSoundEventName:
			key = 109
		}
		newData[key] = value
	}
	return newData
}

// upgradeKey upgrades the legacy key of an entity metadata map to the latest key.
func upgradeKey(data map[uint32]any) map[uint32]any {
	newData := make(map[uint32]any)
	for key, value := range data {
		switch key {
		case 60:
			key = protocol.EntityDataKeyDataRadius
		case 61:
			key = protocol.EntityDataKeyDataWaiting
		case 62:
			key = protocol.EntityDataKeyDataParticle
		case 64:
			key = protocol.EntityDataKeyAttachFace
		case 66:
			key = protocol.EntityDataKeyAttachedPosition
		case 67:
			key = protocol.EntityDataKeyTradeTarget
		case 70:
			key = protocol.EntityDataKeyCommandName
		case 71:
			key = protocol.EntityDataKeyLastCommandOutput
		case 72:
			key = protocol.EntityDataKeyTrackCommandOutput
		case 73:
			key = protocol.EntityDataKeyControllingSeatIndex
		case 74:
			key = protocol.EntityDataKeyStrength
		case 75:
			key = protocol.EntityDataKeyStrengthMax
		case 77:
			key = protocol.EntityDataKeyDataLifetimeTicks
		case 78:
			key = protocol.EntityDataKeyPoseIndex
		case 79:
			key = protocol.EntityDataKeyDataTickOffset
		case 80:
			key = protocol.EntityDataKeyAlwaysShowNameTag
		case 81:
			key = protocol.EntityDataKeyColorTwoIndex
		case 83:
			key = protocol.EntityDataKeyScore
		case 84:
			key = protocol.EntityDataKeyBalloonAnchor
		case 85:
			key = protocol.EntityDataKeyPuffedState
		case 86:
			key = protocol.EntityDataKeyBubbleTime
		case 87:
			key = protocol.EntityDataKeyAgent
		case 90:
			key = protocol.EntityDataKeyEatingCounter
		case 91:
			key = protocol.EntityDataKeyFlagsTwo
		case 94:
			key = protocol.EntityDataKeyDataDuration
		case 95:
			key = protocol.EntityDataKeyDataSpawnTime
		case 96:
			key = protocol.EntityDataKeyDataChangeRate
		case 97:
			key = protocol.EntityDataKeyDataChangeOnPickup
		case 98:
			key = protocol.EntityDataKeyDataPickupCount
		case 99:
			key = protocol.EntityDataKeyInteractText
		case 100:
			key = protocol.EntityDataKeyTradeTier
		case 101:
			key = protocol.EntityDataKeyMaxTradeTier
		case 102:
			key = protocol.EntityDataKeyTradeExperience
		case 104:
			key = protocol.EntityDataKeySkinID
		case 105:
			key = protocol.EntityDataKeyCommandBlockTickDelay
		case 106:
			key = protocol.EntityDataKeyCommandBlockExecuteOnFirstTick
		case 107:
			key = protocol.EntityDataKeyAmbientSoundInterval
		case 108:
			key = protocol.EntityDataKeyAmbientSoundIntervalRange
		case 109:
			key = protocol.EntityDataKeyAmbientSoundEventName
		}
		newData[key] = value
	}
	return newData
}
