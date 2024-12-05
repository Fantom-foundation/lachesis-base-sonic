package abft

import (
	"math/rand"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
	"github.com/Fantom-foundation/lachesis-base/utils/adapters"
	"github.com/Fantom-foundation/lachesis-base/vecfc"
)

type dbEvent struct {
	hash        ltypes.EventHash
	validatorId ltypes.ValidatorID
	seq         ltypes.EventID
	frame       ltypes.FrameID
	lamportTs   ltypes.Lamport
	parents     []ltypes.EventHash
}

type applyBlockFn func(block *ltypes.Block) *ltypes.Validators

type BlockKey struct {
	Epoch ltypes.EpochID
	Frame ltypes.FrameID
}

type BlockResult struct {
	Atropos    ltypes.EventHash
	Cheaters   ltypes.Cheaters
	Validators *ltypes.Validators
}

// CoreLachesis extends Indexed Orderer for tests.
type CoreLachesis struct {
	*IndexedLachesis

	blocks      map[BlockKey]*BlockResult
	lastBlock   BlockKey
	epochBlocks map[ltypes.EpochID]ltypes.FrameID

	applyBlock applyBlockFn
}

// NewCoreLachesis creates empty abft consensus with mem store and optional node weights w.o. some callbacks usually instantiated by Client
func NewCoreLachesis(nodes []ltypes.ValidatorID, weights []ltypes.Weight, mods ...memorydb.Mod) (*CoreLachesis, *Store, *EventStore, *adapters.VectorToDagIndexer) {
	validators := make(ltypes.ValidatorsBuilder, len(nodes))
	for i, v := range nodes {
		if weights == nil {
			validators[v] = 1
		} else {
			validators[v] = weights[i]
		}
	}

	openEDB := func(epoch ltypes.EpochID) kvdb.Store {
		return memorydb.New()
	}
	crit := func(err error) {
		panic(err)
	}
	store := NewStore(memorydb.New(), openEDB, crit, LiteStoreConfig())

	err := store.ApplyGenesis(&Genesis{
		Validators: validators.Build(),
		Epoch:      FirstEpoch,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore()

	config := LiteConfig()
	dagIndexer := &adapters.VectorToDagIndexer{Index: vecfc.NewIndex(crit, vecfc.LiteConfig())}
	lch := NewIndexedLachesis(store, input, dagIndexer, crit, config)

	extended := &CoreLachesis{
		IndexedLachesis: lch,
		blocks:          map[BlockKey]*BlockResult{},
		epochBlocks:     map[ltypes.EpochID]ltypes.FrameID{},
	}

	err = extended.Bootstrap(ltypes.ConsensusCallbacks{
		BeginBlock: func(block *ltypes.Block) ltypes.BlockCallbacks {
			return ltypes.BlockCallbacks{
				EndBlock: func() (sealEpoch *ltypes.Validators) {
					// track blocks
					key := BlockKey{
						Epoch: extended.store.GetEpoch(),
						Frame: extended.store.GetLastDecidedFrame() + 1,
					}
					extended.blocks[key] = &BlockResult{
						Atropos:    block.Atropos,
						Cheaters:   block.Cheaters,
						Validators: extended.store.GetValidators(),
					}
					// check that prev block exists
					if extended.lastBlock.Epoch != key.Epoch && key.Frame != 1 {
						panic("first frame must be 1")
					}
					extended.epochBlocks[key.Epoch]++
					extended.lastBlock = key
					if extended.applyBlock != nil {
						return extended.applyBlock(block)
					}
					return nil
				},
			}
		},
	})
	if err != nil {
		panic(err)
	}

	return extended, store, input, dagIndexer
}

func mutateValidators(validators *ltypes.Validators) *ltypes.Validators {
	r := rand.New(rand.NewSource(int64(validators.TotalWeight()))) // nolint:gosec
	builder := ltypes.NewBuilder()
	for _, vid := range validators.IDs() {
		stake := uint64(validators.Get(vid))*uint64(500+r.Intn(500))/1000 + 1
		builder.Set(vid, ltypes.Weight(stake))
	}
	return builder.Build()
}

// EventStore is a abft event storage for test purpose.
// It implements EventSource interface.
type EventStore struct {
	db map[ltypes.EventHash]ltypes.Event
}

// NewEventStore creates store over memory map.
func NewEventStore() *EventStore {
	return &EventStore{
		db: map[ltypes.EventHash]ltypes.Event{},
	}
}

// Close leaves underlying database.
func (s *EventStore) Close() {
	s.db = nil
}

// SetEvent stores event.
func (s *EventStore) SetEvent(e ltypes.Event) {
	s.db[e.ID()] = e
}

// GetEvent returns stored event.
func (s *EventStore) GetEvent(h ltypes.EventHash) ltypes.Event {
	return s.db[h]
}

// HasEvent returns true if event exists.
func (s *EventStore) HasEvent(h ltypes.EventHash) bool {
	_, ok := s.db[h]
	return ok
}
