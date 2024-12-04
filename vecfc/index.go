package vecfc

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/Fantom-foundation/lachesis-base/utils/simplewlru"
	"github.com/Fantom-foundation/lachesis-base/vecengine"
)

// IndexCacheConfig - config for cache sizes of Engine
type IndexCacheConfig struct {
	ForklessCausePairs   int
	HighestBeforeSeqSize uint
	LowestAfterSeqSize   uint
}

// IndexConfig - Engine config (cache sizes)
type IndexConfig struct {
	Caches IndexCacheConfig
}

// Index is a data to detect forkless-cause condition, calculate median timestamp, detect forks.
type Index struct {
	*vecengine.Engine

	crit          func(error)
	validators    *ltypes.Validators
	validatorIdxs map[ltypes.ValidatorID]ltypes.ValidatorIdx

	getEvent func(ltypes.EventHash) ltypes.Event

	vecDb kvdb.Store
	table struct {
		HighestBeforeSeq kvdb.Store `table:"S"`
		LowestAfterSeq   kvdb.Store `table:"s"`
	}

	cache struct {
		HighestBeforeSeq *simplewlru.Cache
		LowestAfterSeq   *simplewlru.Cache
		ForklessCause    *simplewlru.Cache
	}

	cfg IndexConfig
}

// DefaultConfig returns default index config
func DefaultConfig(scale cachescale.Func) IndexConfig {
	return IndexConfig{
		Caches: IndexCacheConfig{
			ForklessCausePairs:   scale.I(20000),
			HighestBeforeSeqSize: scale.U(160 * 1024),
			LowestAfterSeqSize:   scale.U(160 * 1024),
		},
	}
}

// LiteConfig returns default index config for tests
func LiteConfig() IndexConfig {
	return DefaultConfig(cachescale.Ratio{Base: 100, Target: 1})
}

// NewIndex creates Index instance.
func NewIndex(crit func(error), config IndexConfig) *Index {
	vi := &Index{
		cfg:  config,
		crit: crit,
	}
	vi.Engine = vecengine.NewIndex(crit, vi.GetEngineCallbacks())
	vi.initCaches()

	return vi
}

func NewIndexWithEngine(crit func(error), config IndexConfig, engine *vecengine.Engine) *Index {
	vi := &Index{
		Engine: engine,
		cfg:    config,
		crit:   crit,
	}
	vi.initCaches()

	return vi
}

func (vi *Index) initCaches() {
	vi.cache.ForklessCause, _ = simplewlru.New(uint(vi.cfg.Caches.ForklessCausePairs), vi.cfg.Caches.ForklessCausePairs)
	vi.cache.HighestBeforeSeq, _ = simplewlru.New(vi.cfg.Caches.HighestBeforeSeqSize, int(vi.cfg.Caches.HighestBeforeSeqSize))
	vi.cache.LowestAfterSeq, _ = simplewlru.New(vi.cfg.Caches.LowestAfterSeqSize, int(vi.cfg.Caches.HighestBeforeSeqSize))
}

// Reset resets buffers.
func (vi *Index) Reset(validators *ltypes.Validators, db kvdb.FlushableKVStore, getEvent func(ltypes.EventHash) ltypes.Event) {
	vi.Engine.Reset(validators, db, getEvent)
	vi.vecDb = db
	table.MigrateTables(&vi.table, vi.vecDb)
	vi.getEvent = getEvent
	vi.validators = validators
	vi.validatorIdxs = validators.Idxs()
	vi.cache.ForklessCause.Purge()
	vi.onDropNotFlushed()
}

func (vi *Index) GetEngineCallbacks() vecengine.Callbacks {
	return vecengine.Callbacks{
		GetHighestBefore: func(event ltypes.EventHash) vecengine.HighestBeforeI {
			return vi.GetHighestBefore(event)
		},
		GetLowestAfter: func(event ltypes.EventHash) vecengine.LowestAfterI {
			return vi.GetLowestAfter(event)
		},
		SetHighestBefore: func(event ltypes.EventHash, b vecengine.HighestBeforeI) {
			vi.SetHighestBefore(event, b.(*HighestBeforeSeq))
		},
		SetLowestAfter: func(event ltypes.EventHash, b vecengine.LowestAfterI) {
			vi.SetLowestAfter(event, b.(*LowestAfterSeq))
		},
		NewHighestBefore: func(size ltypes.ValidatorIdx) vecengine.HighestBeforeI {
			return NewHighestBeforeSeq(size)
		},
		NewLowestAfter: func(size ltypes.ValidatorIdx) vecengine.LowestAfterI {
			return NewLowestAfterSeq(size)
		},
		OnDropNotFlushed: vi.onDropNotFlushed,
	}
}

func (vi *Index) onDropNotFlushed() {
	vi.cache.HighestBeforeSeq.Purge()
	vi.cache.LowestAfterSeq.Purge()
}

// GetMergedHighestBefore returns HighestBefore vector clock without branches, where branches are merged into one
func (vi *Index) GetMergedHighestBefore(id ltypes.EventHash) *HighestBeforeSeq {
	return vi.Engine.GetMergedHighestBefore(id).(*HighestBeforeSeq)
}
