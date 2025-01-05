package legacyprotocol

import (
	"math/big"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BitSet ...
func BitSet(v uint64, size int) protocol.Bitset { // this is from @dasciam, :>
	bitset := protocol.NewBitset(size)
	b := big.NewInt(int64(v))

	for s := range min(size, 64) {
		if b.Bit(s) != 0 {
			bitset.Set(s)
		}
	}
	return bitset
}
