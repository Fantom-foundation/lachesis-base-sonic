package tdag

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type TestEventMarshaling struct {
	Epoch ltypes.EpochID
	Seq   ltypes.EventID

	Frame ltypes.FrameID

	Creator ltypes.ValidatorID

	Parents ltypes.EventHashes

	Lamport ltypes.Lamport

	ID   ltypes.EventHash
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
