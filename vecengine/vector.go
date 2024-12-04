package vecengine

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type LowestAfterI interface {
	InitWithEvent(i idx.ValidatorIdx, e ltypes.Event)
	Visit(i idx.ValidatorIdx, e ltypes.Event) bool
}

type HighestBeforeI interface {
	InitWithEvent(i idx.ValidatorIdx, e ltypes.Event)
	IsEmpty(i idx.ValidatorIdx) bool
	IsForkDetected(i idx.ValidatorIdx) bool
	Seq(i idx.ValidatorIdx) idx.EventID
	MinSeq(i idx.ValidatorIdx) idx.EventID
	SetForkDetected(i idx.ValidatorIdx)
	CollectFrom(other HighestBeforeI, branches idx.ValidatorIdx)
	GatherFrom(to idx.ValidatorIdx, other HighestBeforeI, from []idx.ValidatorIdx)
}

type allVecs struct {
	after  LowestAfterI
	before HighestBeforeI
}
