package abft

import (
	"fmt"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/utils/adapters"
	"github.com/Fantom-foundation/lachesis-base/vector"
)

type applyBlockFn func(block *lachesis.Block) *pos.Validators

// TestLachesis extends Lachesis for tests.
type TestLachesis struct {
	*Lachesis

	blocks map[idx.Block]*lachesis.Block

	applyBlock applyBlockFn
}

// FakeLachesis creates empty abft with mem store and equal stakes of nodes in genesis.
func FakeLachesis(nodes []idx.StakerID, stakes []pos.Stake, mods ...memorydb.Mod) (*TestLachesis, *Store, *EventStore) {
	validators := make(pos.ValidatorsBuilder, len(nodes))
	for i, v := range nodes {
		if stakes == nil {
			validators[v] = 1
		} else {
			validators[v] = stakes[i]
		}
	}

	mems := memorydb.NewProducer("", mods...)
	openEDB := func(epoch idx.Epoch) kvdb.DropableStore {
		return mems.OpenDb(fmt.Sprintf("test%d", epoch))
	}
	crit := func(err error) {
		panic(err)
	}
	store := NewStore(mems.OpenDb("test"), openEDB, crit, LiteStoreConfig())

	err := store.ApplyGenesis(&Genesis{
		Validators: validators.Build(),
		Atropos:    hash.ZeroEvent,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore()

	config := LiteConfig()
	lch := NewLachesis(store, input, &adapters.VectorToDagIndexer{vector.NewIndex(crit, vector.LiteConfig())}, crit, config)

	extended := &TestLachesis{
		Lachesis: lch,
		blocks:   map[idx.Block]*lachesis.Block{},
	}

	blockIdx := idx.Block(0)

	err = extended.Bootstrap(lachesis.ConsensusCallbacks{
		BeginBlock: func(block *lachesis.Block) lachesis.BlockCallbacks {
			blockIdx++
			return lachesis.BlockCallbacks{
				EndBlock: func() (sealEpoch *pos.Validators) {
					// track blocks
					extended.blocks[blockIdx] = block
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

	return extended, store, input
}