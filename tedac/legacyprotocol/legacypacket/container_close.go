package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ContainerClose is sent by the server to close a container the player currently has opened, which was opened
// using the ContainerOpen packet, or by the client to tell the server it closed a particular container, such
// as the crafting grid.
type ContainerClose struct {
	// WindowID is the ID representing the window of the container that should be closed. It must be equal to
	// the one sent in the ContainerOpen packet to close the designated window.
	WindowID byte
	// ServerSide determines whether or not the container was force-closed by the server. If this value is
	// not set correctly, the client may ignore the packet and respond with a PacketViolationWarning.
	ServerSide bool
}

// ID ...
func (*ContainerClose) ID() uint32 {
	return packet.IDContainerClose
}

// Marshal ...
func (pk *ContainerClose) Marshal(io protocol.IO) {
	io.Uint8(&pk.WindowID)
	io.Bool(&pk.ServerSide)
}
