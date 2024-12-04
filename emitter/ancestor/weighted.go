package ancestor

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
)

type Metric uint64

type MetricStrategy struct {
	metricFn func(hash.EventHashes) Metric
}

func NewMetricStrategy(metricFn func(hash.EventHashes) Metric) *MetricStrategy {
	return &MetricStrategy{metricFn}
}

// Choose chooses the hash from the specified options
func (st *MetricStrategy) Choose(existing hash.EventHashes, options hash.EventHashes) int {
	var maxI int
	var maxWeight Metric
	// find option with a maximum weight
	for i, opt := range options {
		weight := st.metricFn(append(existing.Copy(), opt))
		if maxWeight == 0 || weight > maxWeight {
			maxI = i
			maxWeight = weight
		}
	}
	return maxI
}
