package chunk

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
)

// NetworkDecode decodes the network serialised data passed into a Chunk if successful. If not, the chunk
// returned is nil and the error non-nil.
// The sub chunk count passed must be that found in the LevelChunk packet.
// noinspection GoUnusedExportedFunction
func NetworkDecode(air uint32, buf *bytes.Buffer, count int, oldFormat bool, r cube.Range) (*Chunk, error) {
	var (
		c   = New(air, r)
		err error
	)
	for i := 0; i < count; i++ {
		index := uint8(i)
		if oldFormat {
			index += 4
		}
		c.sub[index], err = DecodeSubChunk(air, r, buf, &index, NetworkEncoding)
		if err != nil {
			return nil, err
		}
	}
	if oldFormat {
		// Read the old biomes.
		biomes := make([]byte, 256)
		if _, err := buf.Read(biomes); err != nil {
			return nil, fmt.Errorf("error reading biomes: %w", err)
		}

		// Make our 2D biomes 3D.
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				id := biomes[(x&15)|(z&15)<<4]
				for y := r.Min(); y <= r.Max(); y++ {
					c.SetBiome(uint8(x), int16(y), uint8(z), uint32(id))
				}
			}
		}
	} else {
		var last *PalettedStorage
		for i := 0; i < len(c.sub); i++ {
			b, err := decodePalettedStorage(buf, NetworkEncoding, BiomePaletteEncoding)
			if err != nil {
				return nil, err
			}
			// b == nil means this paletted storage had the flag pointing to the previous one. It basically means we should
			// inherit whatever palette we decoded last.
			if i == 0 && b == nil {
				// This should never happen and there is no way to handle this.
				return nil, fmt.Errorf("first biome storage pointed to previous one")
			}
			if b == nil {
				// This means this paletted storage had the flag pointing to the previous one. It basically means we should
				// inherit whatever palette we decoded last.
				b = last
			} else {
				last = b
			}
			c.biomes[i] = b
		}
	}
	return c, nil
}

// DecodeSubChunk decodes a SubChunk from a bytes.Buffer. The Encoding passed defines how the block storages of the
// SubChunk are decoded.
func DecodeSubChunk(air uint32, r cube.Range, buf *bytes.Buffer, index *byte, e Encoding) (*SubChunk, error) {
	ver, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading version: %w", err)
	}
	sub := NewSubChunk(air)
	switch ver {
	default:
		return nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 1:
		// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
		storage, err := decodePalettedStorage(buf, e, BlockPaletteEncoding)
		if err != nil {
			return nil, err
		}
		sub.storages = append(sub.storages, storage)
	case 8, 9:
		// Version 8 allows up to 256 layers for one sub chunk.
		storageCount, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("error reading storage count: %w", err)
		}
		if ver == 9 {
			uIndex, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading subchunk index: %w", err)
			}
			// The index as written here isn't the actual index of the subchunk within the chunk. Rather, it is the Y
			// value of the subchunk. This means that we need to translate it to an index.
			*index = uint8(int8(uIndex) - int8(r[0]>>4))
		}
		sub.storages = make([]*PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.storages[i], err = decodePalettedStorage(buf, e, BlockPaletteEncoding)
			if err != nil {
				return nil, err
			}
		}
	}
	return sub, nil
}

// decodePalettedStorage decodes a PalettedStorage from a bytes.Buffer. The Encoding passed is used to read either a
// network or disk block storage.
func decodePalettedStorage(buf *bytes.Buffer, e Encoding, pe paletteEncoding) (*PalettedStorage, error) {
	blockSize, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading block size: %w", err)
	}
	if e == NetworkEncoding && blockSize&1 != 1 {
		e = NetworkPersistentEncoding
	}

	blockSize >>= 1
	if blockSize == 0x7f {
		return nil, nil
	}

	size := paletteSize(blockSize)
	uint32Count := size.uint32s()

	uint32s := make([]uint32, uint32Count)
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v) %T: not enough block data present: expected %v bytes, got %v", blockSize, pe, byteCount, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	p, err := e.decodePalette(buf, paletteSize(blockSize), pe)
	return newPalettedStorage(uint32s, p), err
}
