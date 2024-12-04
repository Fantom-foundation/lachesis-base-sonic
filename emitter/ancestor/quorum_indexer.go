package ancestor

import (
	"math"
	"sort"

	"github.com/Fantom-foundation/lachesis-base/abft/dagidx"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
	"github.com/Fantom-foundation/lachesis-base/utils/wmedian"
)

type DagIndexQ interface {
	dagidx.VectorClock
}
type DiffMetricFn func(median, current, update idx.EventID, validatorIdx idx.ValidatorIdx) Metric

type QuorumIndexer struct {
	dagi       DagIndexQ
	validators *ltypes.Validators

	globalMatrix     Matrix
	selfParentSeqs   []idx.EventID
	globalMedianSeqs []idx.EventID
	dirty            bool
	searchStrategy   SearchStrategy

	diffMetricFn DiffMetricFn
}

func NewQuorumIndexer(validators *ltypes.Validators, dagi DagIndexQ, diffMetricFn DiffMetricFn) *QuorumIndexer {
	return &QuorumIndexer{
		globalMatrix:     NewMatrix(validators.Len(), validators.Len()),
		globalMedianSeqs: make([]idx.EventID, validators.Len()),
		selfParentSeqs:   make([]idx.EventID, validators.Len()),
		dagi:             dagi,
		validators:       validators,
		diffMetricFn:     diffMetricFn,
		dirty:            true,
	}
}

type Matrix struct {
	buffer  []idx.EventID
	columns idx.ValidatorIdx
}

func NewMatrix(rows, cols idx.ValidatorIdx) Matrix {
	return Matrix{
		buffer:  make([]idx.EventID, rows*cols),
		columns: cols,
	}
}

func (m Matrix) Row(i idx.ValidatorIdx) []idx.EventID {
	return m.buffer[i*m.columns : (i+1)*m.columns]
}

func (m Matrix) Clone() Matrix {
	buffer := make([]idx.EventID, len(m.buffer))
	copy(buffer, m.buffer)
	return Matrix{
		buffer,
		m.columns,
	}
}

func seqOf(seq dagidx.Seq) idx.EventID {
	if seq.IsForkDetected() {
		return math.MaxUint32/2 - 1
	}
	return seq.Seq()
}

type weightedSeq struct {
	seq    idx.EventID
	weight ltypes.Weight
}

func (ws weightedSeq) Weight() ltypes.Weight {
	return ws.weight
}

func (h *QuorumIndexer) ProcessEvent(event ltypes.Event, selfEvent bool) {
	vecClock := h.dagi.GetMergedHighestBefore(event.ID())
	creatorIdx := h.validators.GetIdx(event.Creator())
	// update global matrix
	for validatorIdx := idx.ValidatorIdx(0); validatorIdx < h.validators.Len(); validatorIdx++ {
		seq := seqOf(vecClock.Get(validatorIdx))
		h.globalMatrix.Row(validatorIdx)[creatorIdx] = seq
		if selfEvent {
			h.selfParentSeqs[validatorIdx] = seq
		}
	}
	h.dirty = true
}

func (h *QuorumIndexer) recacheState() {
	// update median seqs
	for validatorIdx := idx.ValidatorIdx(0); validatorIdx < h.validators.Len(); validatorIdx++ {
		pairs := make([]wmedian.WeightedValue, h.validators.Len())
		for i := range pairs {
			pairs[i] = weightedSeq{
				seq:    h.globalMatrix.Row(validatorIdx)[i],
				weight: h.validators.GetWeightByIdx(idx.ValidatorIdx(i)),
			}
		}
		sort.Slice(pairs, func(i, j int) bool {
			a, b := pairs[i].(weightedSeq), pairs[j].(weightedSeq)
			return a.seq > b.seq
		})
		median := wmedian.Of(pairs, h.validators.Quorum())
		h.globalMedianSeqs[validatorIdx] = median.(weightedSeq).seq
	}
	h.searchStrategy = NewMetricStrategy(h.GetMetricOf)
	h.dirty = false
}

func (h *QuorumIndexer) GetMetricOf(parents hash.EventHashes) Metric {
	if h.dirty {
		h.recacheState()
	}
	vecClock := make([]dagidx.HighestBeforeSeq, len(parents))
	for i, parent := range parents {
		vecClock[i] = h.dagi.GetMergedHighestBefore(parent)
	}
	var metric Metric
	for validatorIdx := idx.ValidatorIdx(0); validatorIdx < h.validators.Len(); validatorIdx++ {

		//find the Highest of all the parents
		var update idx.EventID
		for i, _ := range parents {
			if seqOf(vecClock[i].Get(validatorIdx)) > update {
				update = seqOf(vecClock[i].Get(validatorIdx))
			}
		}
		current := h.selfParentSeqs[validatorIdx]
		median := h.globalMedianSeqs[validatorIdx]
		metric += h.diffMetricFn(median, current, update, validatorIdx)
	}
	return metric
}

func (h *QuorumIndexer) SearchStrategy() SearchStrategy {
	if h.dirty {
		h.recacheState()
	}
	return h.searchStrategy
}

func (h *QuorumIndexer) GetGlobalMedianSeqs() []idx.EventID {
	if h.dirty {
		h.recacheState()
	}
	return h.globalMedianSeqs
}

func (h *QuorumIndexer) GetGlobalMatrix() Matrix {
	return h.globalMatrix
}

func (h *QuorumIndexer) GetSelfParentSeqs() []idx.EventID {
	return h.selfParentSeqs
}
