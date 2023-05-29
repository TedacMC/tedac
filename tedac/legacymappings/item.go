package legacymappings

import (
	_ "embed"
	"encoding/json"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

var (
	//go:embed required_item_list.json
	itemIDData []byte

	// items holds a list of all existing items in the game.
	items []protocol.ItemEntry
	// itemIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemIDsToNames = map[int16]string{}
	// itemNamesToIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToIDs = map[string]int16{}
)

// init reads all item entries from the resource JSON, and sets the according values in the maps.
func init() {
	var m map[string]struct {
		RuntimeID      int16 `json:"runtime_id"`
		ComponentBased bool  `json:"component_based"`
	}
	if err := json.Unmarshal(itemIDData, &m); err != nil {
		panic(err)
	}
	for name, data := range m {
		items = append(items, protocol.ItemEntry{Name: name, RuntimeID: data.RuntimeID, ComponentBased: data.ComponentBased})
		itemNamesToIDs[name] = data.RuntimeID
		itemIDsToNames[data.RuntimeID] = name
	}
}

// ItemNameByID returns an item's name by its legacy ID.
func ItemNameByID(id int16) (string, bool) {
	// TODO: Properly handle item aliases.
	name, ok := itemIDsToNames[id]
	return name, ok
}

// ItemIDByName returns an item's ID by its name.
func ItemIDByName(name string) (int16, bool) {
	// TODO: Properly handle item aliases.
	id, ok := itemNamesToIDs[name]
	if !ok {
		id = itemNamesToIDs["minecraft:name_tag"]
	}
	return id, ok
}

// Items returns a slice of all item entries.
func Items() []protocol.ItemEntry {
	return items
}
