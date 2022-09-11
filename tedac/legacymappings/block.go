package legacymappings

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

var (
	//go:embed block_id_map.json
	blockIDData []byte
	//go:embed required_block_states.json
	requiredBlockData []byte

	// blocks holds a list of all existing v in the game.
	blocks []legacyprotocol.BlockEntry
	// legacyToRuntimeIDs maps a legacy block ID to a runtime ID.
	legacyToRuntimeIDs = map[int16]int32{}
	// runtimeToLegacyIDs maps a runtime ID to a legacy block ID.
	runtimeToLegacyIDs = map[int32]int16{}
)

// init reads all block entries from the resource JSON, and sets the according values in the maps.
func init() {
	var legacyIDs map[string]int16
	if err := json.Unmarshal(blockIDData, &legacyIDs); err != nil {
		panic(err)
	}

	var requiredBlockStates map[string]map[string][]int16
	if err := json.Unmarshal(requiredBlockData, &requiredBlockStates); err != nil {
		panic(err)
	}

	for prefix, entries := range requiredBlockStates {
		for identifier, states := range entries {
			name := fmt.Sprintf("%v:%v", prefix, identifier)
			legacyID := legacyIDs[name]
			for _, state := range states {
				blocks = append(blocks, legacyprotocol.BlockEntry{
					Name:     name,
					Data:     state,
					LegacyID: legacyID,
				})
			}
		}
	}

	for runtimeID, entry := range blocks {
		if entry.Data > 15 {
			// TODO: Support data values bigger than 4 bits.
			continue
		}

		ind := (entry.LegacyID << 4) | entry.Data
		legacyToRuntimeIDs[ind] = int32(runtimeID)
		runtimeToLegacyIDs[int32(runtimeID)] = ind
	}
}

// Blocks returns a slice of all block entries.
func Blocks() []legacyprotocol.BlockEntry {
	return blocks
}
