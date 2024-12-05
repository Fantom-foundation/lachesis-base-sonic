package ltypes

import (
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/ethereum/go-ethereum/rlp"
)

type (
	cache struct {
		indexes     map[ValidatorID]ValidatorIdx
		weights     []Weight
		ids         []ValidatorID
		totalWeight Weight
	}
	// Validators group of an epoch with weights.
	// Optimized for BFT algorithm calculations.
	// Read-only.
	Validators struct {
		values map[ValidatorID]Weight
		cache  cache
	}

	// ValidatorsBuilder is a helper to create Validators object
	ValidatorsBuilder map[ValidatorID]Weight
)

// NewBuilder creates new mutable ValidatorsBuilder
func NewBuilder() ValidatorsBuilder {
	return ValidatorsBuilder{}
}

// Set appends item to ValidatorsBuilder object
func (vv ValidatorsBuilder) Set(id ValidatorID, weight Weight) {
	if weight == 0 {
		delete(vv, id)
	} else {
		vv[id] = weight
	}
}

// Build new read-only Validators object
func (vv ValidatorsBuilder) Build() *Validators {
	return newValidators(vv)
}

// EqualWeightValidators builds new read-only Validators object with equal weights (for tests)
func EqualWeightValidators(ids []ValidatorID, weight Weight) *Validators {
	builder := NewBuilder()
	for _, id := range ids {
		builder.Set(id, weight)
	}
	return builder.Build()
}

// ArrayToValidators builds new read-only Validators object from array
func ArrayToValidators(ids []ValidatorID, weights []Weight) *Validators {
	builder := NewBuilder()
	for i, id := range ids {
		builder.Set(id, weights[i])
	}
	return builder.Build()
}

// newValidators builds new read-only Validators object
func newValidators(values ValidatorsBuilder) *Validators {
	valuesCopy := make(ValidatorsBuilder)
	for id, s := range values {
		valuesCopy.Set(id, s)
	}

	vv := &Validators{
		values: valuesCopy,
	}
	vv.cache = vv.calcCaches()
	return vv
}

// Len returns count of validators in Validators objects
func (vv *Validators) Len() ValidatorIdx {
	return ValidatorIdx(len(vv.values))
}

// calcCaches calculates internal caches for validators
func (vv *Validators) calcCaches() cache {
	cache := cache{
		indexes: make(map[ValidatorID]ValidatorIdx),
		weights: make([]Weight, vv.Len()),
		ids:     make([]ValidatorID, vv.Len()),
	}

	for i, v := range vv.sortedArray() {
		cache.indexes[v.ID] = ValidatorIdx(i)
		cache.weights[i] = v.Weight
		cache.ids[i] = v.ID
		totalWeightBefore := cache.totalWeight
		cache.totalWeight += v.Weight
		// check overflow
		if cache.totalWeight < totalWeightBefore {
			panic("validators weight overflow")
		}
	}
	if cache.totalWeight > math.MaxUint32/2 {
		panic("validators weight overflow")
	}

	return cache
}

// get returns weight for validator by ID
func (vv *Validators) Get(id ValidatorID) Weight {
	return vv.values[id]
}

// GetIdx returns index (offset) of validator in the group
func (vv *Validators) GetIdx(id ValidatorID) ValidatorIdx {
	return vv.cache.indexes[id]
}

// GetID returns index validator ID by index (offset) of validator in the group
func (vv *Validators) GetID(i ValidatorIdx) ValidatorID {
	return vv.cache.ids[i]
}

// GetWeightByIdx returns weight for validator by index
func (vv *Validators) GetWeightByIdx(i ValidatorIdx) Weight {
	return vv.cache.weights[i]
}

// Exists returns boolean true if address exists in Validators object
func (vv *Validators) Exists(id ValidatorID) bool {
	_, ok := vv.values[id]
	return ok
}

// IDs returns not sorted ids.
func (vv *Validators) IDs() []ValidatorID {
	return vv.cache.ids
}

// SortedIDs returns deterministically sorted ids.
// The order is the same as for Idxs().
func (vv *Validators) SortedIDs() []ValidatorID {
	return vv.cache.ids
}

// SortedWeights returns deterministically sorted weights.
// The order is the same as for Idxs().
func (vv *Validators) SortedWeights() []Weight {
	return vv.cache.weights
}

// Idxs gets deterministic total order of validators.
func (vv *Validators) Idxs() map[ValidatorID]ValidatorIdx {
	return vv.cache.indexes
}

// sortedArray is sorted by weight and ID
func (vv *Validators) sortedArray() validators {
	array := make(validators, 0, len(vv.values))
	for id, s := range vv.values {
		array = append(array, validator{
			ID:     id,
			Weight: s,
		})
	}
	sort.Sort(array)
	return array
}

// Copy constructs a copy.
func (vv *Validators) Copy() *Validators {
	return newValidators(vv.values)
}

// Builder returns a mutable copy of content
func (vv *Validators) Builder() ValidatorsBuilder {
	return vv.Copy().values
}

// Quorum limit of validators.
func (vv *Validators) Quorum() Weight {
	return vv.TotalWeight()*2/3 + 1
}

// TotalWeight of validators.
func (vv *Validators) TotalWeight() (sum Weight) {
	return vv.cache.totalWeight
}

// EncodeRLP is for RLP serialization.
func (vv *Validators) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, vv.sortedArray())
}

// DecodeRLP is for RLP deserialization.
func (vv *Validators) DecodeRLP(s *rlp.Stream) error {
	var arr []validator
	if err := s.Decode(&arr); err != nil {
		return err
	}

	builder := NewBuilder()
	for _, w := range arr {
		builder.Set(w.ID, w.Weight)
	}
	*vv = *builder.Build()

	return nil
}

func (vv *Validators) String() string {
	str := ""
	for i, vid := range vv.SortedIDs() {
		if len(str) != 0 {
			str += ","
		}
		str += fmt.Sprintf("[%d:%d]", vid, vv.GetWeightByIdx(ValidatorIdx(i)))
	}
	return str
}
