package legacyprotocol

import "github.com/sandertv/gophertunnel/minecraft/protocol"

func ByteSlice(io protocol.IO, x []byte) {
	IoBackwardsCompatibility(io, func(reader *protocol.Reader) {
		ReadByteSlice(reader, x)
	}, func(writer *protocol.Writer) {
		WriteByteSlice(writer, x)
	})
}

func ReadByteSlice(r *protocol.Reader, x []byte) {
	var dataLen uint32
	r.Uint32(&dataLen)
	x = make([]byte, dataLen)
	r.Bytes(&x)
}

func WriteByteSlice(w *protocol.Writer, x []byte) {
	dataLen := uint32(len(x))
	w.Uint32(&dataLen)
	w.Bytes(&x)
}
