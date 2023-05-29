package legacymappings

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockEntry is a block sent in the StartGame packet block runtime ID table. It holds a name and a metadata
// value of a block.
type BlockEntry struct {
	// Name is the name of the custom block.
	Name string
	// Properties is a list of properties which, in combination with the name, specify a unique block.
	Properties map[string]any
}

// Marshal ...
func (b BlockEntry) Marshal(r protocol.IO) {
	r.String(&b.Name)
	r.NBT(&b.Properties, nbt.NetworkLittleEndian)
}

// ItemEntry is an item sent in the StartGame item table. It holds a name and a legacy ID, which is used to
// point back to that name.
type ItemEntry struct {
	// Name if the name of the item, which is a name like 'minecraft:stick'.
	Name string
	// RuntimeID is the ID that is used to identify the item over network. After sending all items in the
	// StartGame packet, items will then be identified using these numerical IDs.
	RuntimeID int16
	// ComponentBased specifies if the item was created using components, meaning the item is a custom item.
	ComponentBased bool
}

// Marshal ...
func (i ItemEntry) Marshal(r protocol.IO) {
	r.String(&i.Name)
	r.Int16(&i.RuntimeID)
	r.Bool(&i.ComponentBased)
}
