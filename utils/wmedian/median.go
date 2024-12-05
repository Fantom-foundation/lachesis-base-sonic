package wmedian

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type WeightedValue interface {
	Weight() ltypes.Weight
}

func Of(values []WeightedValue, stop ltypes.Weight) WeightedValue {
	// Calculate weighted median
	var curWeight ltypes.Weight
	for _, value := range values {
		curWeight += value.Weight()
		if curWeight >= stop {
			return value
		}
	}
	panic("invalid median")
}
