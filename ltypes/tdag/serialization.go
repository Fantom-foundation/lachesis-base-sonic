package tdag

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type TestEventMarshaling struct {
	Epoch idx.EpochID
	Seq   idx.EventID

	Frame idx.FrameID

	Creator idx.ValidatorID

	Parents hash.EventHashes

	Lamport idx.Lamport

	ID   hash.EventHash
	Name string
}

// EventToBytes serializes events
func (e *TestEvent) Bytes() []byte {
	b, _ := rlp.EncodeToBytes(&TestEventMarshaling{
		Epoch:   e.Epoch(),
		Seq:     e.Seq(),
		Frame:   e.Frame(),
		Creator: e.Creator(),
		Parents: e.Parents(),
		Lamport: e.Lamport(),
		ID:      e.ID(),
		Name:    e.Name,
	})
	return b
}
