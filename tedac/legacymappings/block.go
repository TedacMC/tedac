package legacymappings

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/tedacmc/tedac/tedac/latestmappings"
	"github.com/tedacmc/tedac/tedac/legacychunk"
)

var (
	//go:embed block_id_map.json
	blockIDData []byte
	//go:embed 1.12.0_to_1.18.10_blockstate_map.bin
	blockStateMap []byte

	// blocks holds a list of all existing v in the game.
	blocks []BlockEntry

	// stateToRuntimeID maps a block state hash to a runtime ID.
	stateToRuntimeID = map[latestmappings.StateHash]uint32{}
	// runtimeIDToState maps a runtime ID to a state.
	runtimeIDToState = map[uint32]blockupgrader.BlockState{}
)

// init reads all block entries from the resource JSON, and sets the according values in the maps.
func init() {
	var legacyIDs map[string]int16
	if err := json.Unmarshal(blockIDData, &legacyIDs); err != nil {
		panic(err)
	}

	buf := protocol.NewReader(bytes.NewBuffer(blockStateMap), 0, false)
	var length uint32
	buf.Varuint32(&length)
	for i := uint32(0); i < length; i++ {
		var legacyStringId string
		buf.String(&legacyStringId)

		var pairs uint32
		buf.Varuint32(&pairs)
		for y := uint32(0); y < pairs; y++ {
			var meta uint32
			buf.Varuint32(&meta)

			var blockStateRaw map[string]any
			buf.NBT(&blockStateRaw, nbt.LittleEndian)
			latestBlockState := blockupgrader.Upgrade(blockupgrader.BlockState{
				Name:       blockStateRaw["name"].(string),
				Properties: blockStateRaw["states"].(map[string]any),
				Version:    blockStateRaw["version"].(int32),
			})
			legacyId, _ := legacyIDs[legacyStringId]
			legacyRID := uint32(len(blocks))
			blocks = append(blocks, BlockEntry{
				Name:     legacyStringId,
				Data:     int16(meta),
				LegacyID: legacyId,
			})
			stateToRuntimeID[latestmappings.HashState(latestBlockState)] = legacyRID
			runtimeIDToState[legacyRID] = latestBlockState
		}
	}
}

// StateToRuntimeID converts a name and its state properties to a runtime ID.
func StateToRuntimeID(name string, properties map[string]any) uint32 {
	rid, ok := stateToRuntimeID[latestmappings.HashState(blockupgrader.Upgrade(blockupgrader.BlockState{
		Name:       name,
		Properties: properties,
		Version:    legacychunk.LegacyBlockVersion,
	}))]
	if !ok {
		rid = stateToRuntimeID[latestmappings.HashState(blockupgrader.BlockState{Name: "minecraft:info_update"})]
	}
	return rid
}

// RuntimeIDToState converts a runtime ID to a name and its state properties.
func RuntimeIDToState(runtimeID uint32) (name string, properties map[string]any, found bool) {
	s := runtimeIDToState[runtimeID]
	return s.Name, s.Properties, true
}

// Blocks returns a slice of all block entries.
func Blocks() []BlockEntry {
	return blocks
}
