package legacyprotocol

import "github.com/sandertv/gophertunnel/minecraft/protocol"

func IoBackwardsCompatibility(io protocol.IO, readerFunc func(reader *protocol.Reader), writerFunc func(*protocol.Writer)) {
	// I couldn't be bothered to figure out how I could make IO work in the "correct" way
	switch p := io.(type) {
	case *protocol.Reader:
		readerFunc(p)
	case *protocol.Writer:
		writerFunc(p)
	default:
		panic("io isn't recognised")
	}
}
