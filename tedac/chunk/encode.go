package chunk

import (
	"bytes"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// pool is used to pool byte buffers used for encoding chunks.
var pool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// EncodeSubChunk encodes a sub-chunk from a chunk into bytes. An Encoding may be passed to encode either for network or
// disk purposed, the most notable difference being that the network encoding generally uses varints and no NBT.
func EncodeSubChunk(s *SubChunk, e Encoding, r cube.Range, ind int) []byte {
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	_, _ = buf.Write([]byte{SubChunkVersion, byte(len(s.storages)), uint8(ind + (r[0] >> 4))})
	for _, storage := range s.storages {
		encodePalettedStorage(buf, storage, e, BlockPaletteEncoding)
	}
	sub := make([]byte, buf.Len())
	_, _ = buf.Read(sub)
	return sub
}

// encodePalettedStorage encodes a PalettedStorage into a bytes.Buffer. The Encoding passed is used to write the Palette
// of the PalettedStorage.
func encodePalettedStorage(buf *bytes.Buffer, storage *PalettedStorage, e Encoding, pe paletteEncoding) {
	b := make([]byte, len(storage.indices)*4+1)
	b[0] = byte(storage.bitsPerIndex<<1) | e.network()

	for i, v := range storage.indices {
		// Explicitly don't use the binary package to greatly improve performance of writing the uint32s.
		b[i*4+1], b[i*4+2], b[i*4+3], b[i*4+4] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	_, _ = buf.Write(b)

	e.encodePalette(buf, storage.palette, pe)
}
