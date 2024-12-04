package abft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/lachesis-base/abft/dagidx"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

var _ lachesis.Consensus = (*IndexedLachesis)(nil)

// IndexedLachesis performs events ordering and detects cheaters
// It's a wrapper around Orderer, which adds features which might potentially be application-specific:
// confirmed events traversal, DAG index updates and cheaters detection.
// Use this structure if need a general-purpose consensus. Instead, use lower-level abft.Orderer.
type IndexedLachesis struct {
	*Lachesis
	dagIndexer    DagIndexer
	uniqueDirtyID uniqueID
}

type DagIndexer interface {
	dagidx.VectorClock
	dagidx.ForklessCause

	Add(ltypes.Event) error
	Flush()
	DropNotFlushed()

	Reset(validators *ltypes.Validators, db kvdb.FlushableKVStore, getEvent func(hash.Event) ltypes.Event)
}

// NewIndexedLachesis creates IndexedLachesis instance.
func NewIndexedLachesis(store *Store, input EventSource, dagIndexer DagIndexer, crit func(error), config Config) *IndexedLachesis {
	p := &IndexedLachesis{
		Lachesis:      NewLachesis(store, input, dagIndexer, crit, config),
		dagIndexer:    dagIndexer,
		uniqueDirtyID: uniqueID{new(big.Int)},
	}

	return p
}

// Build fills consensus-related fields: Frame, IsRoot
// returns error if event should be dropped
func (p *IndexedLachesis) Build(e ltypes.MutableEvent) error {
	e.SetID(p.uniqueDirtyID.sample())

	defer p.dagIndexer.DropNotFlushed()
	err := p.dagIndexer.Add(e)
	if err != nil {
		return err
	}

	return p.Lachesis.Build(e)
}

// Process takes event into processing.
// Event order matter: parents first.
// All the event checkers must be launched.
// Process is not safe for concurrent use.
func (p *IndexedLachesis) Process(e ltypes.Event) (err error) {
	defer p.dagIndexer.DropNotFlushed()
	err = p.dagIndexer.Add(e)
	if err != nil {
		return err
	}

	err = p.Lachesis.Process(e)
	if err != nil {
		return err
	}
	p.dagIndexer.Flush()
	return nil
}

func (p *IndexedLachesis) Bootstrap(callback lachesis.ConsensusCallbacks) error {
	base := p.Lachesis.OrdererCallbacks()
	ordererCallbacks := OrdererCallbacks{
		ApplyAtropos: base.ApplyAtropos,
		EpochDBLoaded: func(epoch idx.EpochID) {
			if base.EpochDBLoaded != nil {
				base.EpochDBLoaded(epoch)
			}
			p.dagIndexer.Reset(p.store.GetValidators(), flushable.Wrap(p.store.epochTable.VectorIndex), p.input.GetEvent)
		},
	}
	return p.Lachesis.BootstrapWithOrderer(callback, ordererCallbacks)
}

type uniqueID struct {
	counter *big.Int
}

func (u *uniqueID) sample() [24]byte {
	u.counter = u.counter.Add(u.counter, common.Big1)
	var id [24]byte
	copy(id[:], u.counter.Bytes())
	return id
}
