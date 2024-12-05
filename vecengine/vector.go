package vecengine

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type LowestAfterI interface {
	InitWithEvent(i ltypes.ValidatorIdx, e ltypes.Event)
	Visit(i ltypes.ValidatorIdx, e ltypes.Event) bool
}

type HighestBeforeI interface {
	InitWithEvent(i ltypes.ValidatorIdx, e ltypes.Event)
	IsEmpty(i ltypes.ValidatorIdx) bool
	IsForkDetected(i ltypes.ValidatorIdx) bool
	Seq(i ltypes.ValidatorIdx) ltypes.EventID
	MinSeq(i ltypes.ValidatorIdx) ltypes.EventID
	SetForkDetected(i ltypes.ValidatorIdx)
	CollectFrom(other HighestBeforeI, branches ltypes.ValidatorIdx)
	GatherFrom(to ltypes.ValidatorIdx, other HighestBeforeI, from []ltypes.ValidatorIdx)
}

type allVecs struct {
	after  LowestAfterI
	before HighestBeforeI
}
