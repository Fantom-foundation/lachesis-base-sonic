package tdag

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type TestEvent struct {
	ltypes.MutableBaseEvent
	Name string
}

func (e *TestEvent) AddParent(id ltypes.EventHash) {
	parents := e.Parents()
	parents.Add(id)
	e.SetParents(parents)
}
