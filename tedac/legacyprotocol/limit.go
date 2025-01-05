package legacyprotocol

import (
	"fmt"
	"math"
)

const lowerLimit = 64
const mediumLimit = 256
const higherLimit = 1024

// LimitUint32 checks if the value passed is lower than the limit passed. If not, the Reader panics.
func LimitUint32(value uint32, max uint32) {
	if max == math.MaxUint32 {
		// Account for 0-1 overflowing into max.
		max = 0
	}
	if value > max {
		panic(fmt.Sprintf("uint32 %v exceeds maximum of %v", value, max))
	}
}

// LimitInt32 checks if the value passed is lower than the limit passed and higher than the minimum. If not,
// the Reader panics.
func LimitInt32(value int32, min, max int32) {
	if value < min {
		panic(fmt.Sprintf("int32 %v exceeds minimum of %v", value, min))
	} else if value > max {
		panic(fmt.Sprintf("int32 %v exceeds maximum of %v", value, max))
	}
}
