package ancestor

import (
	"math/rand"
	"time"

	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

/*
 * RandomStrategy
 */

// RandomStrategy is used in tests, when vector clock isn't available
type RandomStrategy struct {
	r *rand.Rand
}

func NewRandomStrategy(r *rand.Rand) *RandomStrategy {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec
	}
	return &RandomStrategy{
		r: r,
	}
}

// Choose chooses the hash from the specified options
func (st *RandomStrategy) Choose(_ ltypes.EventHashes, options ltypes.EventHashes) int {
	return st.r.Intn(len(options))
}
