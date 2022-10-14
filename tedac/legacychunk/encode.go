package legacychunk

import (
	"bytes"
	"sync"
)

const (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written on disk and over network.
	SubChunkVersion = 8
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	CurrentBlockVersion int32 = 17825806
)

// pool is used to pool byte buffers used for encoding chunks.
var pool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// SerialisedData holds the serialised data of a chunk. It consists of the chunk's block data itself, a height map, the
// biomes and entities and block entities.
type SerialisedData struct {
	// sub holds the data of the serialised sub chunks in a chunk. Sub chunks that are empty or that otherwise
	// don't exist are represented as an empty slice (or technically, nil).
	SubChunks [][]byte
	// Data2D is the 2D data of the chunk, which is composed of the biome IDs (256 bytes) and optionally the
	// height map of the chunk.
	Data2D []byte
	// BlockNBT is an encoded NBT array of all blocks that carry additional NBT, such as chests, with all
	// their contents.
	BlockNBT []byte
}

// Encode encodes Chunk to an intermediate representation SerialisedData. An Encoding may be passed to encode either for
// network or disk purposed, the most notable difference being that the network encoding generally uses varints and no
// NBT.
func Encode(c *Chunk, e Encoding) SerialisedData {
	d := SerialisedData{SubChunks: make([][]byte, len(c.sub))}
	for i, s := range c.sub {
		d.SubChunks[i] = EncodeSubChunk(s, e)
	}
	d.Data2D = e.data2D(c)
	return d
}

// EncodeSubChunk encodes a sub-chunk from a chunk into bytes. An Encoding may be passed to encode either for network or
// disk purposed, the most notable difference being that the network encoding generally uses varints and no NBT.
func EncodeSubChunk(s *SubChunk, e Encoding) []byte {
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	_, _ = buf.Write([]byte{SubChunkVersion, byte(len(s.storages))})
	for _, storage := range s.storages {
		encodeBlockStorage(buf, storage, e)
	}

	sub := make([]byte, buf.Len())
	_, _ = buf.Read(sub)
	return sub
}

// encodeBlockStorage encodes a BlockStorage into a bytes.Buffer. The Encoding passed is used to write the Palette of
// the BlockStorage.
func encodeBlockStorage(buf *bytes.Buffer, storage *BlockStorage, e Encoding) {
	b := make([]byte, len(storage.blocks)*4+1)
	b[0] = byte(storage.bitsPerBlock<<1) | e.network()

	for i, v := range storage.blocks {
		// Explicitly don't use the binary package to greatly improve performance of writing the uint32s.
		b[i*4+1], b[i*4+2], b[i*4+3], b[i*4+4] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	_, _ = buf.Write(b)

	e.encodePalette(buf, storage.palette)
}
