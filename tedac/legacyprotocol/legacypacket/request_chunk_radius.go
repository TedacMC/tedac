package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RequestChunkRadius is sent by the client to the server to update the server on the chunk view radius that
// it has set in the settings. The server may respond with a ChunkRadiusUpdated packet with either the chunk
// radius requested, or a different chunk radius if the server chooses so.
type RequestChunkRadius struct {
	// ChunkRadius is the requested chunk radius. This value is always the value set in the settings of the
	// player.
	ChunkRadius int32
}

// ID ...
func (*RequestChunkRadius) ID() uint32 {
	return packet.IDRequestChunkRadius
}

func (pk *RequestChunkRadius) Marshal(io protocol.IO) {
	io.Varint32(&pk.ChunkRadius)
}
