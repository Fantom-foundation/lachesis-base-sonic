package vecengine

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// BranchesInfo contains information about global branches of each validator
type BranchesInfo struct {
	BranchIDLastSeq     []ltypes.EventID        // branchID -> highest e.Seq in the branch
	BranchIDCreatorIdxs []ltypes.ValidatorIdx   // branchID -> validator idx
	BranchIDByCreators  [][]ltypes.ValidatorIdx // validator idx -> list of branch IDs
}

// InitBranchesInfo loads BranchesInfo from store
func (vi *Engine) InitBranchesInfo() {
	if vi.bi == nil {
		// if not cached
		vi.bi = vi.getBranchesInfo()
		if vi.bi == nil {
			// first run
			vi.bi = newInitialBranchesInfo(vi.validators)
		}
	}
}

func newInitialBranchesInfo(validators *ltypes.Validators) *BranchesInfo {
	branchIDCreators := validators.SortedIDs()
	branchIDCreatorIdxs := make([]ltypes.ValidatorIdx, len(branchIDCreators))
	for i := range branchIDCreators {
		branchIDCreatorIdxs[i] = ltypes.ValidatorIdx(i)
	}

	branchIDLastSeq := make([]ltypes.EventID, len(branchIDCreatorIdxs))
	branchIDByCreators := make([][]ltypes.ValidatorIdx, validators.Len())
	for i := range branchIDByCreators {
		branchIDByCreators[i] = make([]ltypes.ValidatorIdx, 1, validators.Len()/2+1)
		branchIDByCreators[i][0] = ltypes.ValidatorIdx(i)
	}
	return &BranchesInfo{
		BranchIDLastSeq:     branchIDLastSeq,
		BranchIDCreatorIdxs: branchIDCreatorIdxs,
		BranchIDByCreators:  branchIDByCreators,
	}
}

func (vi *Engine) AtLeastOneFork() bool {
	return ltypes.ValidatorIdx(len(vi.bi.BranchIDCreatorIdxs)) > vi.validators.Len()
}

func (vi *Engine) BranchesInfo() *BranchesInfo {
	return vi.bi
}
