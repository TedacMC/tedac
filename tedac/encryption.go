package tedac

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// cfb holds an encryption session with several fields required to encryption and/or decrypt incoming
// packets.
type cfb struct {
	sendCounter int64
	keyBytes    []byte
	cipherBlock cipher.Block
	iv          []byte
}

// newCFBEncryption returns a new encryption 'session' using the secret key bytes passed. The session has its cipher
// block and IV prepared so that it may be used to decrypt and encrypt data.
func newCFBEncryption(keyBytes []byte) *cfb {
	block, _ := aes.NewCipher(keyBytes[:])
	return &cfb{
		keyBytes:    keyBytes,
		cipherBlock: block,
		iv:          append([]byte(nil), keyBytes[:aes.BlockSize]...),
	}
}

// Encrypt ...
func (c *cfb) Encrypt(data []byte) []byte {
	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, c.sendCounter)
	c.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(buf.Bytes()[:8])
	hash.Write(data[1:])
	hash.Write(c.keyBytes[:])

	// We add the first 8 bytes of the checksum to the data and encrypt it.
	data = append(data, hash.Sum(nil)[:8]...)

	// We skip the very first byte as it contains the header which we need to not encrypt.
	for i := range data[:len(data)-1] {
		offset := i + 1
		// We have to create a new CFBEncrypter for each byte that we decrypt, as this is CFB8.
		encrypter := cipher.NewCFBEncrypter(c.cipherBlock, c.iv)
		encrypter.XORKeyStream(data[offset:offset+1], data[offset:offset+1])
		// For each byte we encrypt, we need to update the IV we have. Each byte encrypted is added to the end
		// of the IV so that the first byte of the IV 'falls off'.
		c.iv = append(c.iv[1:], data[offset])
	}
	return data
}

// Decrypt ...
func (c *cfb) Decrypt(data []byte) {
	for offset, b := range data {
		// Create a new CFBDecrypter for each byte, as we're dealing with CFB8 and have a new IV after each
		// byte that we decrypt.
		decrypter := cipher.NewCFBDecrypter(c.cipherBlock, c.iv)
		decrypter.XORKeyStream(data[offset:offset+1], data[offset:offset+1])

		// Each byte that we decrypt should be added to the end of the IV so that the first byte 'falls off'.
		c.iv = append(c.iv[1:], b)
	}
}

// Verify ...
func (c *cfb) Verify(data []byte) error {
	sum := data[len(data)-8:]

	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, c.sendCounter)
	c.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(buf.Bytes())
	hash.Write(data[:len(data)-8])
	hash.Write(c.keyBytes[:])
	ourSum := hash.Sum(nil)[:8]

	// Finally we check if the original sum was equal to the sum we just produced.
	if !bytes.Equal(sum, ourSum) {
		return fmt.Errorf("invalid packet checksum: %v should be %v", hex.EncodeToString(sum), hex.EncodeToString(ourSum))
	}
	return nil
}
