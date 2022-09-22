package latestmappings

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/segmentio/fasthash/fnv1"
	"sort"
	"strings"
	"unsafe"
)

// State holds a combination of a name and properties, together with a version.
type State struct {
	// Name is the name of the block.
	Name string `nbt:"name"`
	// Properties is a map of properties that define the block's state.
	Properties map[string]any `nbt:"states"`
	// Version is the version of the block state.
	Version int32 `nbt:"version"`
}

var (
	//go:embed block_states.nbt
	blockStateData []byte

	// states holds a list of all possible vanilla block states.
	states []State
	// stateRuntimeIDs holds a map for looking up the runtime ID of a block by the stateHash it produces.
	stateRuntimeIDs = map[StateHash]uint32{}
	// runtimeIDToState holds a map for looking up the blockState of a block by its runtime ID.
	runtimeIDToState = map[uint32]State{}
)

var (
	//go:embed item_runtime_ids.nbt
	itemRuntimeIDData []byte
	// itemRuntimeIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemRuntimeIDsToNames = map[int32]string{}
	// itemNamesToRuntimeIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToRuntimeIDs = map[string]int32{}
)

// init initializes the item and state mappings.
func init() {
	var items map[string]int32
	if err := nbt.Unmarshal(itemRuntimeIDData, &items); err != nil {
		panic(err)
	}
	for name, rid := range items {
		itemNamesToRuntimeIDs[name] = rid
		itemRuntimeIDsToNames[rid] = name
	}

	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))

	// Register all block states present in the block_states.nbt file. These are all possible options registered
	// blocks may encode to.
	var s State
	for {
		if err := dec.Decode(&s); err != nil {
			break
		}

		rid := uint32(len(states))
		states = append(states, s)

		stateRuntimeIDs[HashState(s)] = rid
		runtimeIDToState[rid] = s
	}
}

// Adjust adjusts the latest mappings to account for custom states.
func Adjust(customStates []State) {
	adjustedStates := append(states, customStates...)
	sort.SliceStable(adjustedStates, func(i, j int) bool {
		stateOne, stateTwo := adjustedStates[i], adjustedStates[j]
		if stateOne.Name == stateTwo.Name {
			return false
		}
		return fnv1.HashString64(stateOne.Name) < fnv1.HashString64(stateTwo.Name)
	})

	stateRuntimeIDs = make(map[StateHash]uint32, len(adjustedStates))
	runtimeIDToState = make(map[uint32]State, len(adjustedStates))
	for rid, state := range adjustedStates {
		stateRuntimeIDs[HashState(state)] = uint32(rid)
		runtimeIDToState[uint32(rid)] = state
	}
}

// StateToRuntimeID converts a name and its state properties to a runtime ID.
func StateToRuntimeID(name string, properties map[string]any) (runtimeID uint32, found bool) {
	if updated, ok := UpdatedNameFromAlias(name); ok {
		name = updated
	}
	rid, ok := stateRuntimeIDs[HashState(State{Name: name, Properties: properties})]
	return rid, ok
}

// RuntimeIDToState converts a runtime ID to a name and its state properties.
func RuntimeIDToState(runtimeID uint32) (name string, properties map[string]any, found bool) {
	s := runtimeIDToState[runtimeID]
	return s.Name, s.Properties, true
}

// ItemRuntimeIDToName converts an item runtime ID to a string ID.
func ItemRuntimeIDToName(runtimeID int32) (name string, found bool) {
	name, ok := itemRuntimeIDsToNames[runtimeID]
	return name, ok
}

// ItemNameToRuntimeID converts a string ID to an item runtime ID.
func ItemNameToRuntimeID(name string) (runtimeID int32, found bool) {
	rid, ok := itemNamesToRuntimeIDs[name]
	return rid, ok
}

// StateHash is a struct that may be used as a map key for block states. It contains the name of the block state
// and an encoded version of the properties.
type StateHash struct {
	Name, Properties string
}

// HashState produces a hash for the block properties held by the blockState.
func HashState(state State) StateHash {
	if state.Properties == nil {
		return StateHash{Name: state.Name}
	}
	keys := make([]string, 0, len(state.Properties))
	for k := range state.Properties {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var b strings.Builder
	for _, k := range keys {
		switch v := state.Properties[k].(type) {
		case bool:
			if v {
				b.WriteByte(1)
			} else {
				b.WriteByte(0)
			}
		case uint8:
			b.WriteByte(v)
		case int32:
			a := *(*[4]byte)(unsafe.Pointer(&v))
			b.Write(a[:])
		case string:
			b.WriteString(v)
		default:
			// If block encoding is broken, we want to find out as soon as possible. This saves a lot of time
			// debugging in-game.
			panic(fmt.Sprintf("invalid block property type %T for property %v", v, k))
		}
	}
	return StateHash{Name: state.Name, Properties: b.String()}
}
