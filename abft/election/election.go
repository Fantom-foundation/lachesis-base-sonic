package election

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type (
	// Election is cached data of election algorithm.
	Election struct {
		// election params
		frameToDecide ltypes.FrameID

		validators *ltypes.Validators

		// election state
		decidedRoots map[ltypes.ValidatorID]voteValue // decided roots at "frameToDecide"
		votes        map[voteID]voteValue

		// external world
		observe       ForklessCauseFn
		getFrameRoots GetFrameRootsFn
	}

	// ForklessCauseFn returns true if event A is forkless caused by event B
	ForklessCauseFn func(a ltypes.EventHash, b ltypes.EventHash) bool
	// GetFrameRootsFn returns all the roots in the specified frame
	GetFrameRootsFn func(f ltypes.FrameID) []RootAndSlot

	// Slot specifies a root slot {addr, frame}. Normal validators can have only one root with this pair.
	// Due to a fork, different roots may occupy the same slot
	Slot struct {
		Frame     ltypes.FrameID
		Validator ltypes.ValidatorID
	}

	// RootAndSlot specifies concrete root of slot.
	RootAndSlot struct {
		ID   ltypes.EventHash
		Slot Slot
	}
)

type voteID struct {
	fromRoot     RootAndSlot
	forValidator ltypes.ValidatorID
}
type voteValue struct {
	decided      bool
	yes          bool
	observedRoot ltypes.EventHash
}

// Res defines the final election result, i.e. decided frame
type Res struct {
	Frame   ltypes.FrameID
	Atropos ltypes.EventHash
}

// New election context
func New(
	validators *ltypes.Validators,
	frameToDecide ltypes.FrameID,
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
func (el *Election) Reset(validators *ltypes.Validators, frameToDecide ltypes.FrameID) {
	el.validators = validators
	el.frameToDecide = frameToDecide
	el.votes = make(map[voteID]voteValue)
	el.decidedRoots = make(map[ltypes.ValidatorID]voteValue)
}

// return root slots which are not within el.decidedRoots
func (el *Election) notDecidedRoots() []ltypes.ValidatorID {
	notDecidedRoots := make([]ltypes.ValidatorID, 0, el.validators.Len())

	for _, validator := range el.validators.IDs() {
		if _, ok := el.decidedRoots[validator]; !ok {
			notDecidedRoots = append(notDecidedRoots, validator)
		}
	}
	if ltypes.ValidatorIdx(len(notDecidedRoots)+len(el.decidedRoots)) != el.validators.Len() { // sanity check
		panic("Mismatch of roots")
	}
	return notDecidedRoots
}

// observedRoots returns all the roots at the specified frame which do forkless cause the specified root.
func (el *Election) observedRoots(root ltypes.EventHash, frame ltypes.FrameID) []RootAndSlot {
	observedRoots := make([]RootAndSlot, 0, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRoots = append(observedRoots, frameRoot)
		}
	}
	return observedRoots
}

func (el *Election) observedRootsMap(root ltypes.EventHash, frame ltypes.FrameID) map[ltypes.ValidatorID]RootAndSlot {
	observedRootsMap := make(map[ltypes.ValidatorID]RootAndSlot, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRootsMap[frameRoot.Slot.Validator] = frameRoot
		}
	}
	return observedRootsMap
}
