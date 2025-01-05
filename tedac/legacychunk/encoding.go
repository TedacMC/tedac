package legacychunk

import (
	"bytes"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Encoding is an encoding type used for Chunk encoding. Implementations of this interface are DiskEncoding and
// NetworkEncoding, which can be used to encode a Chunk to an intermediate disk or network representation respectively.
type Encoding interface {
	encoding() nbt.Encoding
	encodePalette(buf *bytes.Buffer, p *Palette)
	decodePalette(buf *bytes.Buffer, blockSize paletteSize) (*Palette, error)
	network() byte
	data2D(c *Chunk) []byte
}

// NetworkEncoding is the Encoding used for sending a Chunk over network. It does not use NBT and writes varints.
var NetworkEncoding networkEncoding

// networkEncoding implements the Chunk encoding for sending over network.
type networkEncoding struct{}

func (networkEncoding) network() byte          { return 1 }
func (networkEncoding) encoding() nbt.Encoding { return nbt.NetworkLittleEndian }
func (networkEncoding) data2D(c *Chunk) []byte { return append(c.biomes[:], 0) }
func (networkEncoding) encodePalette(buf *bytes.Buffer, p *Palette) {
	_ = protocol.WriteVarint32(buf, int32(p.Len()))
	for _, runtimeID := range p.blockRuntimeIDs {
		_ = protocol.WriteVarint32(buf, int32(runtimeID))
	}
}
func (networkEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize) (*Palette, error) {
	var paletteCount int32
	if err := protocol.Varint32(buf, &paletteCount); err != nil {
		return nil, fmt.Errorf("error reading palette entry count: %w", err)
	}
	if paletteCount <= 0 {
		return nil, fmt.Errorf("invalid palette entry count %v", paletteCount)
	}

	blocks, temp := make([]uint32, paletteCount), int32(0)
	for i := int32(0); i < paletteCount; i++ {
		if err := protocol.Varint32(buf, &temp); err != nil {
			return nil, fmt.Errorf("error decoding palette entry: %w", err)
		}
		blocks[i] = uint32(temp)
	}
	return &Palette{blockRuntimeIDs: blocks, size: blockSize}, nil
}
