package vecfc

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

func (vi *Index) getBytes(table kvdb.Store, id ltypes.EventHash) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.crit(err)
	}
	return b
}

func (vi *Index) setBytes(table kvdb.Store, id ltypes.EventHash, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.crit(err)
	}
}

// GetLowestAfter reads the vector from DB
func (vi *Index) GetLowestAfter(id ltypes.EventHash) *LowestAfterSeq {
	if bVal, okGet := vi.cache.LowestAfterSeq.Get(id); okGet {
		return bVal.(*LowestAfterSeq)
	}

	b := LowestAfterSeq(vi.getBytes(vi.table.LowestAfterSeq, id))
	if b == nil {
		return nil
	}
	vi.cache.LowestAfterSeq.Add(id, &b, uint(len(b)))
	return &b
}

// GetHighestBefore reads the vector from DB
func (vi *Index) GetHighestBefore(id ltypes.EventHash) *HighestBeforeSeq {
	if bVal, okGet := vi.cache.HighestBeforeSeq.Get(id); okGet {
		return bVal.(*HighestBeforeSeq)
	}

	b := HighestBeforeSeq(vi.getBytes(vi.table.HighestBeforeSeq, id))
	if b == nil {
		return nil
	}
	vi.cache.HighestBeforeSeq.Add(id, &b, uint(len(b)))
	return &b
}

// SetLowestAfter stores the vector into DB
func (vi *Index) SetLowestAfter(id ltypes.EventHash, seq *LowestAfterSeq) {
	vi.setBytes(vi.table.LowestAfterSeq, id, *seq)

	vi.cache.LowestAfterSeq.Add(id, seq, uint(len(*seq)))
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id ltypes.EventHash, seq *HighestBeforeSeq) {
	vi.setBytes(vi.table.HighestBeforeSeq, id, *seq)

	vi.cache.HighestBeforeSeq.Add(id, seq, uint(len(*seq)))
}
