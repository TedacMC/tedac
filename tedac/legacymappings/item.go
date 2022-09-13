package legacymappings

import (
	_ "embed"
	"encoding/json"
)

var (
	//go:embed item_id_map.json
	itemIDData []byte

	// items holds a list of all existing items in the game.
	items []ItemEntry
	// itemIDsToNames holds a map to translate item runtime IDs to string IDs.
	itemIDsToNames = map[int16]string{}
	// itemNamesToIDs holds a map to translate item string IDs to runtime IDs.
	itemNamesToIDs = map[string]int16{}
)

// init reads all item entries from the resource JSON, and sets the according values in the maps.
func init() {
	var m map[string]int16
	if err := json.Unmarshal(itemIDData, &m); err != nil {
		panic(err)
	}
	for name, id := range m {
		items = append(items, ItemEntry{Name: name, LegacyID: id})
		itemNamesToIDs[name] = id
		itemIDsToNames[id] = name
	}
}

// ItemNameByID returns an item's name by its legacy ID.
func ItemNameByID(id int16) (string, bool) {
	name, ok := itemIDsToNames[id]

	if name == "minecraft:netherstar" {
		name = "minecraft:nether_star"
	}
	// TODO: Properly handle item aliases.

	return name, ok
}

// ItemIDByName returns an item's ID by its name.
func ItemIDByName(name string) int16 {
	if name == "minecraft:nether_star" {
		name = "minecraft:netherstar"
	}
	// TODO: Properly handle item aliases.

	id, ok := itemNamesToIDs[name]
	if !ok {
		id = itemNamesToIDs["minecraft:name_tag"]
	}
	return id
}

// Items returns a slice of all item entries.
func Items() []ItemEntry {
	return items
}
