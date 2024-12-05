package tdag

import (
	"strings"

	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// TestEvents is a ordered slice of events.
type TestEvents []*TestEvent

// String returns human readable representation.
func (ee TestEvents) String() string {
	ss := make([]string, len(ee))
	for i := 0; i < len(ee); i++ {
		ss[i] = ee[i].String()
	}
	return strings.Join(ss, " ")
}

// ByParents returns events topologically ordered by parent dependency.
// Used only for tests.
func ByParents(ee ltypes.Events) (res ltypes.Events) {
	unsorted := make(ltypes.Events, len(ee))
	exists := ltypes.EventHashSet{}
	for i, e := range ee {
		unsorted[i] = e
		exists.Add(e.ID())
	}
	ready := ltypes.EventHashSet{}
	for len(unsorted) > 0 {
	EVENTS:
		for i, e := range unsorted {

			for _, p := range e.Parents() {
				if exists.Contains(p) && !ready.Contains(p) {
					continue EVENTS
				}
			}

			res = append(res, e)
			unsorted = append(unsorted[0:i], unsorted[i+1:]...)
			ready.Add(e.ID())
			break
		}
	}

	return
}

// ByParents returns events topologically ordered by parent dependency.
// Used only for tests.
func (ee TestEvents) ByParents() (res TestEvents) {
	unsorted := make(ltypes.Events, len(ee))
	for i, e := range ee {
		unsorted[i] = e
	}
	sorted := ByParents(unsorted)
	res = make(TestEvents, len(ee))
	for i, e := range sorted {
		res[i] = e.(*TestEvent)
	}

	return
}