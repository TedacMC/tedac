package legacymappings

import (
	"bytes"
	_ "embed"

	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/tedacmc/tedac/tedac/latestmappings"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte

	// blocks holds a list of all existing v in the game.
	blocks []protocol.BlockEntry

	// stateToRuntimeID maps a block state hash to a runtime ID.
	stateToRuntimeID = map[latestmappings.StateHash]uint32{}
	// runtimeIDToState maps a runtime ID to a state.
	runtimeIDToState = map[uint32]blockupgrader.BlockState{}
)

// init reads all block entries from the resource JSON, and sets the according values in the maps.
func init() {
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))

	// Register all block states present in the block_states.nbt file. These are all possible options registered
	// blocks may encode to.
	var s blockupgrader.BlockState
	for {
		if err := dec.Decode(&s); err != nil {
			break
		}

		rid := uint32(len(blocks))
		blocks = append(blocks, protocol.BlockEntry{
			Name:       s.Name,
			Properties: s.Properties,
		})
		stateToRuntimeID[latestmappings.HashState(s)] = rid
		runtimeIDToState[rid] = s
	}
}

// StateToRuntimeID converts a name and its state properties to a runtime ID.
func StateToRuntimeID(name string, properties map[string]any) uint32 {
	if alias, ok := latestmappings.AliasFromUpdatedName(name); ok {
		name = alias
	}
	rid, ok := stateToRuntimeID[latestmappings.HashState(blockupgrader.BlockState{Name: name, Properties: properties})]
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
func Blocks() []protocol.BlockEntry {
	return blocks
}
