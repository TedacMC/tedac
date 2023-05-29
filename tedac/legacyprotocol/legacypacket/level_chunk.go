package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// LevelChunk is sent by the server to provide the client with a chunk of a world data (16xYx16 blocks).
// Typically, a certain amount of chunks is sent to the client before sending it the spawn PlayStatus packet,
// so that the client spawns in a loaded world.
type LevelChunk struct {
	// Position contains the X and Z coordinates of the chunk sent. You can convert a block coordinate to a chunk
	// coordinate by right-shifting it four bits.
	Position protocol.ChunkPos
	// SubChunkCount is the amount of sub chunks that are part of the chunk sent. Depending on if the cache
	// is enabled, a list of blob hashes will be sent, or, if disabled, the sub chunk data.
	SubChunkCount uint32
	// CacheEnabled specifies if the client blob cache should be enabled. This system is based on hashes of
	// blobs which are consistent and saved by the client in combination with that blob, so that the server
	// does not have the same chunk multiple times. If the client does not yet have a blob with the hash sent,
	// it will send a ClientCacheBlobStatus packet containing the hashes is does not have the data of.
	CacheEnabled bool
	// BlobHashes is a list of all blob hashes used in the chunk. It is composed of SubChunkCount + 1 hashes,
	// with the first SubChunkCount hashes being those of the sub chunks and the last one that of the biome
	// of the chunk.
	// If CacheEnabled is set to false, BlobHashes can be left empty.
	BlobHashes []uint64
	// RawPayload is a serialised string of chunk data. The data held depends on if CacheEnabled is set to
	// true. If set to false, the payload is composed of multiple sub-chunks, each of which carry a version
	// which indicates the way they are serialised, followed by biomes, border blocks and tile entities. If
	// CacheEnabled is true, the payload consists out of the border blocks and tile entities only.
	RawPayload []byte
}

// ID ...
func (*LevelChunk) ID() uint32 {
	return packet.IDLevelChunk
}

// Marshal ...
func (pk *LevelChunk) Marshal(w protocol.IO) {
	w.ChunkPos(&pk.Position)
	w.Varuint32(&pk.SubChunkCount)

	w.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		protocol.FuncSlice(w, &pk.BlobHashes, w.Uint64)
	}

	w.ByteSlice(&pk.RawPayload)
}
