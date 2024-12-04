package basestreamseeder

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type testLocator struct {
	B []byte
}

func (l testLocator) Compare(b basestream.Locator) int {
	return bytes.Compare(l.B, b.(testLocator).B)
}

func (l testLocator) Inc() basestream.Locator {
	nextBn := new(big.Int).SetBytes(l.B)
	nextBn.Add(nextBn, common.Big1)
	return testLocator{
		B: nextBn.Bytes(),
	}
}

type testPayload struct {
	IDs    hash.EventHashes
	Events ltypes.Events
	Size   uint64
}

func (p testPayload) AddEvent(id hash.EventHash, event ltypes.Event) {
	p.IDs = append(p.IDs, id)          // nolint:staticcheck
	p.Events = append(p.Events, event) // nolint:staticcheck
	p.Size += uint64(event.Size())     // nolint:staticcheck
}

func (p testPayload) Len() int {
	return len(p.IDs)
}

func (p testPayload) TotalSize() uint64 {
	return p.Size
}

func (p testPayload) TotalMemSize() int {
	return int(p.Size) + len(p.IDs)*128
}
