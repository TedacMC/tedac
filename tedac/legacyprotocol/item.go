package legacyprotocol

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/tedacmc/tedac/tedac/legacymappings"
)

// ItemStack represents an item instance/stack over network. It has a network ID and a metadata value that
// define its type.
type ItemStack struct {
	ItemType
	// Count is the count of items that the item stack holds.
	Count int16
	// NBTData is a map that is serialised to its NBT representation when sent in a packet.
	NBTData map[string]any
	// CanBePlacedOn is a list of block identifiers like 'minecraft:stone' which the item, if it is an item
	// that can be placed, can be placed on top of.
	CanBePlacedOn []string
	// CanBreak is a list of block identifiers like 'minecraft:dirt' that the item is able to break.
	CanBreak []string
}

// ItemType represents a consistent combination of network ID and metadata value of an item. It cannot usually
// be changed unless a new item is obtained.
type ItemType struct {
	// NetworkID is the numerical network ID of the item. This is sometimes a positive ID, and sometimes a
	// negative ID, depending on what item it concerns.
	NetworkID int32
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int16
}

// shieldID represents the ID of a shield in the 1.12 item table.
var shieldID, _ = legacymappings.ItemIDByName("minecraft:shield")

// Item see ReadItem and WriteItem for documentation
func Item(io protocol.IO, x *ItemStack) {
	IoBackwardsCompatibility(io, func(reader *protocol.Reader) {
		ReadItem(reader, x)
	}, func(writer *protocol.Writer) {
		WriteItem(writer, x)
	})
}

// ReadItem reads an item stack from buffer src and stores it into item stack x.
func ReadItem(r *protocol.Reader, x *ItemStack) {
	x.NBTData = make(map[string]any)
	r.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there is no more data we should read for the item instance. After all, air
		// items aren't really anything.
		x.MetadataValue, x.Count, x.CanBePlacedOn, x.CanBreak = 0, 0, nil, nil
		return
	}
	var auxValue int32
	r.Varint32(&auxValue)
	x.MetadataValue = int16(auxValue >> 8)
	x.Count = int16(auxValue & 0xff)

	var userDataMarker int16
	r.Int16(&userDataMarker)

	if userDataMarker == -1 {
		var userDataVersion uint8
		r.Uint8(&userDataVersion)

		switch userDataVersion {
		case 1:
			r.NBT(&x.NBTData, nbt.NetworkLittleEndian)
		default:
			r.UnknownEnumOption(userDataVersion, "item user data version")
			return
		}
	} else if userDataMarker > 0 {
		r.NBT(&x.NBTData, nbt.LittleEndian)
	}
	var count int32
	r.Varint32(&count)
	LimitInt32(count, 0, higherLimit)

	x.CanBePlacedOn = make([]string, count)
	for i := int32(0); i < count; i++ {
		r.String(&x.CanBePlacedOn[i])
	}

	r.Varint32(&count)
	LimitInt32(count, 0, higherLimit)

	x.CanBreak = make([]string, count)
	for i := int32(0); i < count; i++ {
		r.String(&x.CanBreak[i])
	}

	if x.NetworkID == int32(shieldID) {
		var blockingTick int64
		r.Varint64(&blockingTick)
	}
}

// WriteItem writes an item stack x to buffer dst.
func WriteItem(w *protocol.Writer, x *ItemStack) {
	w.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there's no more data to follow. Return immediately.
		return
	}
	aux := int32(x.MetadataValue<<8) | int32(x.Count)
	w.Varint32(&aux)
	if len(x.NBTData) != 0 {
		userDataMarker := int16(-1)
		userDataVer := uint8(1)

		w.Int16(&userDataMarker)
		w.Uint8(&userDataVer)
		w.NBT(&x.NBTData, nbt.NetworkLittleEndian)
	} else {
		userDataMarker := int16(0)

		w.Int16(&userDataMarker)
	}
	placeOnLen := int32(len(x.CanBePlacedOn))
	canBreak := int32(len(x.CanBreak))

	w.Varint32(&placeOnLen)
	for _, block := range x.CanBePlacedOn {
		w.String(&block)
	}
	w.Varint32(&canBreak)
	for _, block := range x.CanBreak {
		w.String(&block)
	}

	if x.NetworkID == int32(shieldID) {
		var blockingTick int64
		w.Varint64(&blockingTick)
	}
}
