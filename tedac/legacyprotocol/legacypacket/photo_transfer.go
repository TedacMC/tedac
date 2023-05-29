package legacypacket

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PhotoTransfer is sent by the server to transfer a photo (image) file to the client. It is typically used
// to transfer photos so that the client can display it in a portfolio in Education Edition.
// While previously usable in the default Bedrock Edition, the displaying of photos in books was disabled and
// the packet now has little use anymore.
type PhotoTransfer struct {
	// PhotoName is the name of the photo to transfer. It is the exact file name that the client will download
	// the photo as, including the extension of the file.
	PhotoName string
	// PhotoData is the raw data of the photo image. The format of this data may vary: Formats such as JPEG or
	// PNG work, as long as PhotoName has the correct extension.
	PhotoData []byte
	// BookID is the ID of the book that the photo is associated with. If the PhotoName in a book with this ID
	// is set to PhotoName, it will display the photo (provided Education Edition is used).
	// The photo image is downloaded to a sub-folder with this book ID.
	BookID string
}

// ID ...
func (*PhotoTransfer) ID() uint32 {
	return packet.IDPhotoTransfer
}

// Marshal ...
func (pk *PhotoTransfer) Marshal(w protocol.IO) {
	w.String(&pk.PhotoName)
	w.ByteSlice(&pk.PhotoData)
	w.String(&pk.BookID)
}
