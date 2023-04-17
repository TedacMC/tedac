package legacypacket

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	PlayerListActionAdd = iota
	PlayerListActionRemove
)

// PlayerList is sent by the server to update the client-side player list in the in-game menu screen. It shows
// the icon of each player if the correct XUID is written in the packet.
// Sending the PlayerList packet is obligatory when sending an AddPlayer packet. The added player will not
// show up to a client if it has not been added to the player list, because several properties of the player
// are obtained from the player list, such as the skin.
type PlayerList struct {
	// ActionType is the action to execute upon the player list. The entries that follow specify which entries
	// are added or removed from the player list.
	ActionType byte
	// Entries is a list of all player list entries that should be added/removed from the player list,
	// depending on the ActionType set.
	Entries []PlayerListEntry
}

// PlayerListEntry is an entry found in the PlayerList packet. It represents a single player using the UUID
// found in the entry, and contains several properties such as the skin.
type PlayerListEntry struct {
	// UUID is the UUID of the player as sent in the Login packet when the client joined the server. It must
	// match this UUID exactly for the correct XBOX Live icon to show up in the list.
	UUID uuid.UUID
	// EntityUniqueID is the unique entity ID of the player. This ID typically stays consistent during the
	// lifetime of a world, but servers often send the runtime ID for this.
	EntityUniqueID int64
	// Username is the username that is shown in the player list of the player that obtains a PlayerList
	// packet with this entry. It does not have to be the same as the actual username of the player.
	Username string
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string
	// SkinData is a byte slice of 64*32*4, 64*64*4 or 128*128*4 bytes. It is a RGBA ordered byte
	// representation of the skin colours.
	SkinData []byte
	// CapeData is a byte slice of 64*32*4 bytes. It is a RGBA ordered byte representation of the cape
	// colours, much like the SkinData.
	CapeData []byte
	// SkinGeometryName is the geometry name of the skin geometry above. This name must be equal to one of the
	// outer names found in the SkinGeometry, so that the client can find the correct geometry data.
	SkinGeometryName string
	// SkinGeometry is a base64 JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry []byte
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
}

// ID ...
func (*PlayerList) ID() uint32 {
	return packet.IDPlayerList
}

// Marshal ...
func (pk *PlayerList) Marshal(io protocol.IO) {
	io.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case PlayerListActionAdd:
		protocol.Slice(io, &pk.Entries)
	case PlayerListActionRemove:
		protocol.FuncIOSlice(io, &pk.Entries, PlayerListRemoveEntry)
	default:
		panic("unknown player list action type")
	}
}

// Marshal encodes/decodes a PlayerListEntry.
func (x *PlayerListEntry) Marshal(r protocol.IO) {
	r.UUID(&x.UUID)
	r.Varint64(&x.EntityUniqueID)
	r.String(&x.Username)
	r.String(&x.SkinID)
	r.ByteSlice(&x.SkinData)
	r.ByteSlice(&x.CapeData)
	r.String(&x.SkinGeometryName)
	r.ByteSlice(&x.SkinGeometry)
	r.String(&x.XUID)
	r.String(&x.PlatformChatID)
}

// PlayerListRemoveEntry encodes/decodes a PlayerListEntry for removal from the list.
func PlayerListRemoveEntry(r protocol.IO, x *PlayerListEntry) {
	r.UUID(&x.UUID)
}
