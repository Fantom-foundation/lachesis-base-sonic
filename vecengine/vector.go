package vecengine

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type LowestAfterI interface {
	InitWithEvent(i idx.Validator, e ltypes.Event)
	Visit(i idx.Validator, e ltypes.Event) bool
}

type HighestBeforeI interface {
	InitWithEvent(i idx.Validator, e ltypes.Event)
	IsEmpty(i idx.Validator) bool
	IsForkDetected(i idx.Validator) bool
	Seq(i idx.Validator) idx.EventID
	MinSeq(i idx.Validator) idx.EventID
	SetForkDetected(i idx.Validator)
	CollectFrom(other HighestBeforeI, branches idx.Validator)
	GatherFrom(to idx.Validator, other HighestBeforeI, from []idx.Validator)
}

type allVecs struct {
	after  LowestAfterI
	before HighestBeforeI
}
