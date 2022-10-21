package legacypacket

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
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
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
	// BuildPlatform is the platform of the player as sent by that player in the Login packet.
	BuildPlatform int32
	// Skin is the skin of the player that should be added to the player list. Once sent here, it will not
	// have to be sent again.
	Skin legacyprotocol.Skin
	// Teacher is a Minecraft: Education Edition field. It specifies if the player to be added to the player
	// list is a teacher.
	Teacher bool
	// Host specifies if the player that is added to the player list is the host of the game.
	Host bool
}

// ID ...
func (*PlayerList) ID() uint32 {
	return packet.IDPlayerList
}

// Marshal ...
func (pk *PlayerList) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Entries))
	w.Uint8(&pk.ActionType)
	w.Varuint32(&l)
	for _, entry := range pk.Entries {
		switch pk.ActionType {
		case PlayerListActionAdd:
			w.UUID(&entry.UUID)
			w.Varint64(&entry.EntityUniqueID)
			w.String(&entry.Username)
			w.String(&entry.XUID)
			w.String(&entry.PlatformChatID)
			w.Int32(&entry.BuildPlatform)
			legacyprotocol.WriteSerialisedSkin(w, &entry.Skin)
			w.Bool(&entry.Teacher)
			w.Bool(&entry.Host)
		case PlayerListActionRemove:
			w.UUID(&entry.UUID)
		default:
			panic("unknown player list action type")
		}
	}
	if pk.ActionType == PlayerListActionAdd {
		for _, entry := range pk.Entries {
			w.Bool(&entry.Skin.Trusted)
		}
	}
}

// Unmarshal ...
func (pk *PlayerList) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Uint8(&pk.ActionType)
	r.Varuint32(&count)
	pk.Entries = make([]PlayerListEntry, count)
	for i := uint32(0); i < count; i++ {
		switch pk.ActionType {
		case PlayerListActionAdd:
			r.UUID(&pk.Entries[i].UUID)
			r.Varint64(&pk.Entries[i].EntityUniqueID)
			r.String(&pk.Entries[i].Username)
			r.String(&pk.Entries[i].XUID)
			r.String(&pk.Entries[i].PlatformChatID)
			r.Int32(&pk.Entries[i].BuildPlatform)
			legacyprotocol.SerialisedSkin(r, &pk.Entries[i].Skin)
			r.Bool(&pk.Entries[i].Teacher)
			r.Bool(&pk.Entries[i].Host)
		case PlayerListActionRemove:
			r.UUID(&pk.Entries[i].UUID)
		default:
			panic("unknown player list action type")
		}
	}
	if pk.ActionType == PlayerListActionAdd {
		for i := uint32(0); i < count; i++ {
			r.Bool(&pk.Entries[i].Skin.Trusted)
		}
	}
}
