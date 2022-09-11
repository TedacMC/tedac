package tedac

import (
	"crypto/aes"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Protocol represents the v1.12.0 Protocol implementation.
type Protocol struct{}

// ID ...
func (Protocol) ID() int32 {
	return 361
}

// Ver ...
func (Protocol) Ver() string {
	return "1.12.0"
}

// Packets ...
func (Protocol) Packets() packet.Pool {
	// TODO
	return packet.NewPool()
}

// Encryption ...
func (Protocol) Encryption(key [32]byte) packet.Encryption {
	block, _ := aes.NewCipher(key[:])
	return &cfb{
		keyBytes:    key[:],
		cipherBlock: block,
		iv:          append([]byte(nil), key[:aes.BlockSize]...),
	}
}

// ConvertToLatest ...
func (Protocol) ConvertToLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	// TODO
	return []packet.Packet{pk}
}

// ConvertFromLatest ...
func (Protocol) ConvertFromLatest(pk packet.Packet, _ *minecraft.Conn) []packet.Packet {
	// TODO
	return []packet.Packet{pk}
}
