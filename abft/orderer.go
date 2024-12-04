package abft

import (
	"github.com/Fantom-foundation/lachesis-base/abft/dagidx"
	"github.com/Fantom-foundation/lachesis-base/abft/election"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type OrdererCallbacks struct {
	ApplyAtropos func(decidedFrame ltypes.FrameID, atropos ltypes.EventHash) (sealEpoch *ltypes.Validators)

	EpochDBLoaded func(ltypes.EpochID)
}

type OrdererDagIndex interface {
	dagidx.ForklessCause
}

// Orderer processes events to reach finality on their order.
// Unlike abft.Lachesis, this raw level of abstraction doesn't track cheaters detection
type Orderer struct {
	config Config
	crit   func(error)
	store  *Store
	input  EventSource

	election *election.Election
	dagIndex OrdererDagIndex

	callback OrdererCallbacks
}

// NewOrderer creates Orderer instance.
// Unlike Lachesis, Orderer doesn't updates DAG indexes for events, and doesn't detect cheaters
// It has only one purpose - reaching consensus on events order.
func NewOrderer(store *Store, input EventSource, dagIndex OrdererDagIndex, crit func(error), config Config) *Orderer {
	p := &Orderer{
		config:   config,
		store:    store,
		input:    input,
		crit:     crit,
		dagIndex: dagIndex,
	}

	return p
}
