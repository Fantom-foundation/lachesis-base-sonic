package election

import (
	"errors"
	"github.com/Fantom-foundation/lachesis-base/types"
)

type (
	// Election is cached data of election algorithm.
	Election struct {
		// election params
		frameToDecide types.Frame

		validators *types.Validators

		// election state
		decidedRoots map[types.ValidatorID]voteValue // decided roots at "frameToDecide"
		votes        map[voteID]voteValue

		// external world
		observe       ForklessCauseFn
		getFrameRoots GetFrameRootsFn
	}

	// ForklessCauseFn returns true if event A is forkless caused by event B
	ForklessCauseFn func(a types.Event, b types.Event) bool
	// GetFrameRootsFn returns all the roots in the specified frame
	GetFrameRootsFn func(f types.Frame) []RootAndSlot

	// Slot specifies a root slot {addr, frame}. Normal validators can have only one root with this pair.
	// Due to a fork, different roots may occupy the same slot
	Slot struct {
		Frame     types.Frame
		Validator types.ValidatorID
	}

	// RootAndSlot specifies concrete root of slot.
	RootAndSlot struct {
		ID   types.Event
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
	observedRoot hash.Event
}

// Res defines the final election result, i.e. decided frame
type Res struct {
	Frame   idx.Frame
	Atropos hash.Event
}

// New election context
func New(
	validators *pos.Validators,
	frameToDecide idx.Frame,
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
func (el *Election) Reset(validators *pos.Validators, frameToDecide idx.Frame) {
	el.validators = validators
	el.frameToDecide = frameToDecide
	el.votes = make(map[voteID]voteValue)
	el.decidedRoots = make(map[idx.ValidatorID]voteValue)
}

// return root slots which are not within el.decidedRoots
func (el *Election) notDecidedRoots() []idx.ValidatorID {
	notDecidedRoots := make([]idx.ValidatorID, 0, el.validators.Len())

	for _, validator := range el.validators.IDs() {
<<<<<<< Updated upstream
		if _, ok := el.decidedRoots[validator]; !ok {
			notDecidedRoots = append(notDecidedRoots, validator)
=======
		vote := voteValue{}
		// in initial round, vote "yes" if observe the subject
		observedRoot, ok := observedRootsMap[validator]
		vote.yes = ok
		vote.decided = false
		if ok {
			vote.observedRoot = observedRoot.ID
>>>>>>> Stashed changes
		}
		// save vote for next rounds
		// note that frame is a constant (frame -1 of the root frame)
		vid := voteID{
			fromRoot:     RootAndSlot{rootEvent, Slot{frame, rootValidator}},
			forValidator: validator,
		}
		el.votes[vid] = vote
	}
	if idx.Validator(len(notDecidedRoots)+len(el.decidedRoots)) != el.validators.Len() { // sanity check
		panic("Mismatch of roots")
	}
	return notDecidedRoots
}

// observedRoots returns all the roots at the specified frame which do forkless cause the specified root.
func (el *Election) observedRoots(root hash.Event, frame idx.Frame) []RootAndSlot {
	observedRoots := make([]RootAndSlot, 0, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRoots = append(observedRoots, frameRoot)
		}
	}
	return observedRoots
}

<<<<<<< Updated upstream
func (el *Election) observedRootsMap(root hash.Event, frame idx.Frame) map[idx.ValidatorID]RootAndSlot {
	observedRootsMap := make(map[idx.ValidatorID]RootAndSlot, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRootsMap[frameRoot.Slot.Validator] = frameRoot
=======
func (el *Election) aggregateVotes(frame idx.Frame, rootValidator idx.ValidatorID, rootEvent hash.Event) (*Res, error) {
	var observedRoots []RootAndSlot
	observedRoots = el.observedRoots(rootEvent, frame-1)
	for _, validator := range el.validators.IDs() {
		if _, ok := el.decidedRoots[validator]; !ok {
			yesVotes := el.validators.NewCounter()
			noVotes := el.validators.NewCounter()
			// calc number of "yes" and "no", weighted by validator's weight
			for _, observedRoot := range observedRoots {
				vid := voteID{
					fromRoot:     observedRoot,
					forValidator: validator,
				}
				if vote, ok := el.votes[vid]; ok {
					if vote.yes {
						yesVotes.Count(observedRoot.Slot.Validator)
					} else {
						noVotes.Count(observedRoot.Slot.Validator)
					}
				} else {
					return nil, errors.New("every root must vote for every not decided subject. possibly roots are processed out of order")
				}
			}
			// vote as majority of votes
			vote := voteValue{}
			vote.yes = yesVotes.Sum() >= noVotes.Sum()
			if vote.yes != nil {
				vote.observedRoot = *subjectHash
			}
			// If supermajority is observed, then the final decision may be made.
			// It's guaranteed to be final and consistent unless more than 1/3W are Byzantine.
			vote.decided = yesVotes.HasQuorum() || noVotes.HasQuorum()
			if vote.decided {
				el.decidedRoots[validator] = vote
			}
			// save vote for next rounds
			vid := voteID{
				fromRoot:     RootAndSlot{rootEvent, Slot{frame, rootValidator}},
				forValidator: validator,
			}
			el.votes[vid] = vote
>>>>>>> Stashed changes
		}
	}
	return observedRootsMap
}
