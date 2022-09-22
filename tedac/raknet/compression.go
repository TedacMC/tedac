package raknet

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"io"
	"math"
)

// ZLibCompression is an implementation of the zLib compression algorithm.
type ZLibCompression struct{}

// EncodeCompression ...
func (ZLibCompression) EncodeCompression() uint16 {
	return math.MaxUint16
}

// Compress ...
func (ZLibCompression) Compress(decompressed []byte) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024*1024*2))
	writer := zlib.NewWriter(buf)
	if _, err := writer.Write(decompressed); err != nil {
		return nil, fmt.Errorf("error writing zlib data: %v", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing zlib writer: %v", err)
	}
	return buf.Bytes(), nil
}

// Decompress ...
func (ZLibCompression) Decompress(compressed []byte) ([]byte, error) {
	buf := bytes.NewBuffer(compressed)
	zlibReader, err := zlib.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("error decompressing data: %v", err)
	}
	_ = zlibReader.Close()
	raw, err := io.ReadAll(zlibReader)
	if err != nil {
		return nil, fmt.Errorf("error reading decompressed data: %v", err)
	}
	return raw, nil
}

// init registers the ZLibCompression algorithm.
func init() {
	packet.RegisterCompression(ZLibCompression{})
}
