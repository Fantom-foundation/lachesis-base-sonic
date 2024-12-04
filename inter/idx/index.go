package idx

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
)

type (
	// Epoch numeration.
	EpochID uint32

	// Event numeration.
	EventID uint32

	// Block numeration.
	BlockID uint64

	// Lamport numeration.
	Lamport uint32

	// Frame numeration.
	FrameID uint32

	// Pack numeration.
	Pack uint32

	// ValidatorID numeration.
	ValidatorID uint32
)

// Bytes gets the byte representation of the index.
func (e EpochID) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(e))
}

// Bytes gets the byte representation of the index.
func (e EventID) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(e))
}

// Bytes gets the byte representation of the index.
func (b BlockID) Bytes() []byte {
	return bigendian.Uint64ToBytes(uint64(b))
}

// Bytes gets the byte representation of the index.
func (l Lamport) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(l))
}

// Bytes gets the byte representation of the index.
func (p Pack) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(p))
}

// Bytes gets the byte representation of the index.
func (s ValidatorID) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(s))
}

// Bytes gets the byte representation of the index.
func (f FrameID) Bytes() []byte {
	return bigendian.Uint32ToBytes(uint32(f))
}

// BytesToEpoch converts bytes to epoch index.
func BytesToEpoch(b []byte) EpochID {
	return EpochID(bigendian.BytesToUint32(b))
}

// BytesToEvent converts bytes to event index.
func BytesToEvent(b []byte) EventID {
	return EventID(bigendian.BytesToUint32(b))
}

// BytesToBlock converts bytes to block index.
func BytesToBlock(b []byte) BlockID {
	return BlockID(bigendian.BytesToUint64(b))
}

// BytesToLamport converts bytes to block index.
func BytesToLamport(b []byte) Lamport {
	return Lamport(bigendian.BytesToUint32(b))
}

// BytesToFrame converts bytes to block index.
func BytesToFrame(b []byte) FrameID {
	return FrameID(bigendian.BytesToUint32(b))
}

// BytesToPack converts bytes to block index.
func BytesToPack(b []byte) Pack {
	return Pack(bigendian.BytesToUint32(b))
}

// BytesToValidatorID converts bytes to validator index.
func BytesToValidatorID(b []byte) ValidatorID {
	return ValidatorID(bigendian.BytesToUint32(b))
}

// MaxLamport return max value
func MaxLamport(x, y Lamport) Lamport {
	if x > y {
		return x
	}
	return y
}
