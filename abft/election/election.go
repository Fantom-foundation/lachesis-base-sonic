package election

import (
	"errors"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
)

// Election is cached data of election algorithm.
type Election struct {
	// election params
	frameToDecide idx.Frame

	validators *pos.Validators

	// election state
	decidedRoots map[idx.ValidatorID]voteValue // decided roots at "frameToDecide"
	votes        map[voteID]voteValue

	// external world
	observe       ForklessCauseFn
	getFrameRoots GetFrameRootsFn
}

// ForklessCauseFn returns true if event A is forkless caused by event B
type ForklessCauseFn func(a hash.Event, b hash.Event) bool

// GetFrameRootsFn returns all the roots in the specified frame
type GetFrameRootsFn func(f idx.Frame) []RootAndSlot

// Slot specifies a root slot {addr, frame}. Normal validators can have only one root with this pair.
// Due to a fork, different roots may occupy the same slot
type Slot struct {
	Frame     idx.Frame
	Validator idx.ValidatorID
}

// RootAndSlot specifies concrete root of slot.
type RootAndSlot struct {
	ID   hash.Event
	Slot Slot
}

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

func (el *Election) ProcessRoot(newRoot RootAndSlot) (*Res, error) {
	return el.processRoot(newRoot.Slot.Frame, newRoot.Slot.Validator, newRoot.ID)
}

// ProcessRoot calculates Atropos votes only for the new root.
// If this root observes that the current election is decided, then return decided Atropos
func (el *Election) processRoot(frame idx.Frame, validator idx.ValidatorID, newRoot hash.Event) (*Res, error) {
	round := int32(frame) - int32(el.frameToDecide)
	if round <= 0 {
		// unreachable because of condition above
		return nil, nil
	} else if round == 1 {
		el.performVoting(frame, validator, newRoot)
		return nil, nil
	} else {
		res, err := el.aggregateVotes(frame, validator, newRoot)
		if res != nil || err != nil {
			return res, err
		}
		// check if election is decided
		return el.chooseAtropos()
	}
}

// Chooses the decided "yes" roots with the greatest weight amount.
// This root serves as a "checkpoint" within DAG, as it's guaranteed to be final and consistent unless more than 1/3W are Byzantine.
// Other validators will come to the same Atropos not later than current highest frame + 2.
func (el *Election) chooseAtropos() (*Res, error) {
	// iterate until Yes root is met, which will be Atropos. I.e. not necessarily all the roots must be decided
	for _, validator := range el.validators.SortedIDs() {
		vote, ok := el.decidedRoots[validator]
		if !ok {
			return nil, nil // not decided
		}
		if vote.yes {
			return &Res{
				Frame:   el.frameToDecide,
				Atropos: vote.observedRoot,
			}, nil
		}
	}
	return nil, errors.New("all the roots are decided as 'no', which is possible only if more than 1/3W are Byzantine")
}

func (el *Election) observedRootsMap(event hash.Event, frame idx.Frame) map[idx.ValidatorID]RootAndSlot {
	observedRootsMap := make(map[idx.ValidatorID]RootAndSlot, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(event, frameRoot.ID) {
			observedRootsMap[frameRoot.Slot.Validator] = frameRoot
		}
	}
	return observedRootsMap
}

// performVoting performs voting for the first round
func (el *Election) performVoting(frame idx.Frame, rootValidator idx.ValidatorID, rootEvent hash.Event) {
	var observedRootsMap map[idx.ValidatorID]RootAndSlot
	observedRootsMap = el.observedRootsMap(rootEvent, frame-1)
	for _, validator := range el.validators.IDs() {
		if _, ok := el.decidedRoots[validator]; !ok {
			vote := voteValue{}
			// in initial round, vote "yes" if observe the subject
			observedRoot, ok := observedRootsMap[validator]
			vote.yes = ok
			vote.decided = false
			if ok {
				vote.observedRoot = observedRoot.ID
			}
			// save vote for next rounds
			// note that frame is a constant (frame -1 of the root frame)
			vid := voteID{
				fromRoot:     RootAndSlot{rootEvent, Slot{frame, rootValidator}},
				forValidator: validator,
			}
			el.votes[vid] = vote
		}
	}
}

// observedRoots returns all the roots at the specified frame which do forkless cause the specified root.
func (el *Election) observedRoots(event hash.Event, frame idx.Frame) []RootAndSlot {
	observedRoots := make([]RootAndSlot, 0, el.validators.Len())
	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(event, frameRoot.ID) {
			observedRoots = append(observedRoots, frameRoot)
		}
	}
	return observedRoots
}

func (el *Election) aggregateVotes(frame idx.Frame, rootValidator idx.ValidatorID, rootEvent hash.Event) (*Res, error) {
	var observedRoots []RootAndSlot
	observedRoots = el.observedRoots(rootEvent, frame-1)
	for _, validator := range el.validators.IDs() {
		if _, ok := el.decidedRoots[validator]; !ok {
			yesVotes := el.validators.NewCounter()
			noVotes := el.validators.NewCounter()
			allVotes := el.validators.NewCounter()
			// calc number of "yes" and "no", weighted by validator's weight
			var subjectHash *hash.Event
			for _, observedRoot := range observedRoots {
				vid := voteID{
					fromRoot:     observedRoot,
					forValidator: validator,
				}
				if vote, ok := el.votes[vid]; ok {
					if vote.yes && subjectHash != nil && *subjectHash != vote.observedRoot {
						return nil, fmt.Errorf("forkless caused by 2 fork roots => more than 1/3W are Byzantine (%s != %s, election frame=%d, validator=%d)",
							subjectHash.String(), vote.observedRoot.String(), el.frameToDecide, validator)
					}
					if vote.yes {
						subjectHash = &vote.observedRoot
						yesVotes.Count(observedRoot.Slot.Validator)
					} else {
						noVotes.Count(observedRoot.Slot.Validator)
					}
					if !allVotes.Count(observedRoot.Slot.Validator) {
						// it shouldn't be possible to get here, because we've taken 1 root from every node above
						return nil, fmt.Errorf("forkless caused by 2 fork roots => more than 1/3W are Byzantine (election frame=%d, validator=%d)",
							el.frameToDecide, validator)
					}
				} else {
					return nil, errors.New("every root must vote for every not decided subject. possibly roots are processed out of order")
				}
			}
			// sanity checks
			if !allVotes.HasQuorum() {
				return nil, errors.New("root must be forkless caused by at least 2/3W of prev roots. possibly roots are processed out of order")
			}
			// vote as majority of votes
			vote := voteValue{}
			vote.yes = yesVotes.Sum() >= noVotes.Sum()
			if vote.yes && subjectHash != nil {
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
		}
	}
	return nil, nil
}
