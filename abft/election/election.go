package election

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type (
	// Election is cached data of election algorithm.
	Election struct {
		// election params
		frameToDecide idx.FrameID

		validators *ltypes.Validators

		// election state
		decidedRoots map[idx.ValidatorID]voteValue // decided roots at "frameToDecide"
		votes        map[voteID]voteValue

		// external world
		observe       ForklessCauseFn
		getFrameRoots GetFrameRootsFn
	}

	// ForklessCauseFn returns true if event A is forkless caused by event B
	ForklessCauseFn func(a hash.EventHash, b hash.EventHash) bool
	// GetFrameRootsFn returns all the roots in the specified frame
	GetFrameRootsFn func(f idx.FrameID) []RootAndSlot

	// Slot specifies a root slot {addr, frame}. Normal validators can have only one root with this pair.
	// Due to a fork, different roots may occupy the same slot
	Slot struct {
		Frame     idx.FrameID
		Validator idx.ValidatorID
	}

	// RootAndSlot specifies concrete root of slot.
	RootAndSlot struct {
		ID   hash.EventHash
		Slot Slot
	}
)

type voteID struct {
	fromRoot     RootAndSlot
	forValidator idx.ValidatorID
}
type voteValue struct {
	decided      bool
	yes          bool
	observedRoot hash.EventHash
}

// Res defines the final election result, i.e. decided frame
type Res struct {
	Frame   idx.FrameID
	Atropos hash.EventHash
}

// New election context
func New(
	validators *ltypes.Validators,
	frameToDecide idx.FrameID,
	forklessCauseFn ForklessCauseFn,
	getFrameRoots GetFrameRootsFn,
) *Election {
	el := &Election{
		observe:       forklessCauseFn,
		getFrameRoots: getFrameRoots,
	}

	el.Reset(validators, frameToDecide)

	return el
}

// Reset erases the current election state, prepare for new election frame
func (el *Election) Reset(validators *ltypes.Validators, frameToDecide idx.FrameID) {
	el.validators = validators
	el.frameToDecide = frameToDecide
	el.votes = make(map[voteID]voteValue)
	el.decidedRoots = make(map[idx.ValidatorID]voteValue)
}

// return root slots which are not within el.decidedRoots
func (el *Election) notDecidedRoots() []idx.ValidatorID {
	notDecidedRoots := make([]idx.ValidatorID, 0, el.validators.Len())

	for _, validator := range el.validators.IDs() {
		if _, ok := el.decidedRoots[validator]; !ok {
			notDecidedRoots = append(notDecidedRoots, validator)
		}
	}
	if idx.ValidatorIdx(len(notDecidedRoots)+len(el.decidedRoots)) != el.validators.Len() { // sanity check
		panic("Mismatch of roots")
	}
	return notDecidedRoots
}

// observedRoots returns all the roots at the specified frame which do forkless cause the specified root.
func (el *Election) observedRoots(root hash.EventHash, frame idx.FrameID) []RootAndSlot {
	observedRoots := make([]RootAndSlot, 0, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRoots = append(observedRoots, frameRoot)
		}
	}
	return observedRoots
}

func (el *Election) observedRootsMap(root hash.EventHash, frame idx.FrameID) map[idx.ValidatorID]RootAndSlot {
	observedRootsMap := make(map[idx.ValidatorID]RootAndSlot, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRootsMap[frameRoot.Slot.Validator] = frameRoot
		}
	}
	return observedRootsMap
}
