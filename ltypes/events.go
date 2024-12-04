package ltypes

import (
	"strings"
)

// Events is a ordered slice of events.
type Events []Event

// String returns human readable representation.
func (ee Events) String() string {
	ss := make([]string, len(ee))
	for i := 0; i < len(ee); i++ {
		ss[i] = ee[i].String()
	}
	return strings.Join(ss, " ")
}

func (ee Events) Metric() (metric Metric) {
	metric.Num = EventID(len(ee))
	for _, e := range ee {
		metric.Size += uint64(e.Size())
	}
	return metric
}

func (ee Events) IDs() EventHashes {
	ids := make(EventHashes, len(ee))
	for i, e := range ee {
		ids[i] = e.ID()
	}
	return ids
}
