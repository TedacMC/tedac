package legacymappings

import (
	_ "embed"
	"encoding/json"
	"github.com/tedacmc/tedac/tedac/latestmappings"
)

var (
	//go:embed block_id_map.json
	blockIDData []byte
	//go:embed block_state_meta_map.json
	blockStateMetaData []byte

	// blocks holds a list of all existing v in the game.
	blocks []BlockEntry

	// stateToRuntimeID maps a block state hash to a runtime ID.
	stateToRuntimeID = map[latestmappings.StateHash]uint32{}
	// runtimeIDToState maps a runtime ID to a state.
	runtimeIDToState = map[uint32]latestmappings.State{}
)

// init reads all block entries from the resource JSON, and sets the according values in the maps.
func init() {
	var legacyIDs map[string]int16
	if err := json.Unmarshal(blockIDData, &legacyIDs); err != nil {
		panic(err)
	}

	var blockStateMetas []int16
	if err := json.Unmarshal(blockStateMetaData, &blockStateMetas); err != nil {
		panic(err)
	}

	for latestRID, meta := range blockStateMetas {
		name, properties, _ := latestmappings.RuntimeIDToState(uint32(latestRID))
		if alias, ok := latestmappings.AliasFromUpdatedName(name); ok {
			name = alias
		}

		legacyID, ok := legacyIDs[name]
		if !ok {
			// This block didn't exist in v1.12.0.
			continue
		}

		state := latestmappings.State{Name: name, Properties: properties}
		legacyRID := uint32(len(blocks))

		blocks = append(blocks, BlockEntry{
			Name:     name,
			Data:     meta,
			LegacyID: legacyID,
		})
		stateToRuntimeID[latestmappings.HashState(state)] = legacyRID
		runtimeIDToState[legacyRID] = state
	}
}

// StateToRuntimeID converts a name and its state properties to a runtime ID.
func StateToRuntimeID(name string, properties map[string]any) uint32 {
	if alias, ok := latestmappings.AliasFromUpdatedName(name); ok {
		name = alias
	}
	rid, ok := stateToRuntimeID[latestmappings.HashState(latestmappings.State{Name: name, Properties: properties})]
	if !ok {
		rid = stateToRuntimeID[latestmappings.HashState(latestmappings.State{Name: "minecraft:info_update"})]
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
