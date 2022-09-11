package legacymappings

import (
	_ "embed"
	"encoding/json"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

var (
	//go:embed item_id_map.json
	itemIDData []byte

	// items holds a list of all existing items in the game.
	items []legacyprotocol.ItemEntry
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
		items = append(items, legacyprotocol.ItemEntry{Name: name, LegacyID: id})
		itemNamesToIDs[name] = id
		itemIDsToNames[id] = name
	}
}

// ItemNameByID returns an item's name by its legacy ID.
func ItemNameByID(id int16) (string, bool) {
	name, ok := itemIDsToNames[id]
	return name, ok
}

// ItemIDByName returns an item's ID by its name.
func ItemIDByName(name string) (int16, bool) {
	id, ok := itemNamesToIDs[name]
	return id, ok
}

// Items returns a slice of all item entries.
func Items() []legacyprotocol.ItemEntry {
	return items
}
