package idx

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
)

type (
	// ValidatorIdx numeration.
	ValidatorIdx uint32
)

// Bytes gets the byte representation of the index.
func (v ValidatorIdx) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(v))
}

// BytesToValidator converts bytes to validator index.
func BytesToValidator(b []byte) ValidatorIdx {
	return ValidatorIdx(bigendian.BytesToUint32(b))
}
