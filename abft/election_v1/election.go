package electionv1

import (
	"errors"
	"os"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
)

type (
	ForklessCauseFn func(a hash.Event, b hash.Event) bool
	GetFrameRootsFn func(f idx.Frame) []EventDescriptor
)

type EventDescriptor struct {
	Frame       idx.Frame
	ValidatorID idx.ValidatorID
	EventID     hash.Event
}

type AtroposDecision struct {
	Frame     idx.Frame
	AtroposID hash.Event
}

type Election struct {
	validators *pos.Validators

	forklessCauses ForklessCauseFn
	getFrameRoots  GetFrameRootsFn

	vote     map[idx.Frame]map[hash.Event]map[idx.ValidatorID]bool
	decided  map[idx.Frame]map[idx.ValidatorID]bool
	eventMap map[idx.Frame]map[idx.ValidatorID]hash.Event
	atropos  map[idx.Frame]struct{}
}

// New election context
func New(
	validators *pos.Validators,
	forklessCauseFn ForklessCauseFn,
	getFrameRoots GetFrameRootsFn,
) *Election {
	election := &Election{
		forklessCauses: forklessCauseFn,
		getFrameRoots:  getFrameRoots,
		validators:     validators,
	}
	election.Reset(validators)
	return election
}

func (el *Election) Reset(validators *pos.Validators) {
	el.vote = make(map[idx.Frame]map[hash.Event]map[idx.ValidatorID]bool)
	el.eventMap = make(map[idx.Frame]map[idx.ValidatorID]hash.Event)
	el.decided = make(map[idx.Frame]map[idx.ValidatorID]bool)
	el.atropos = make(map[idx.Frame]struct{})
	el.validators = validators
}

// ProcessRoot calculates Atropos votes only for the new root.
// If this root observes that the current election is decided, then return decided Atropos
func (el *Election) ProcessRoot(
	frame idx.Frame,
	validatorId idx.ValidatorID,
	root hash.Event,
) ([]*AtroposDecision, error) {
	decidedAtropoi := make([]*AtroposDecision, 0)
	// Iterate over all undecided frames
	if _, ok := el.vote[frame]; !ok {
		el.newFrameToDecide(frame)
	}
	for frameToDecide := range el.vote {
		round := int32(frame) - int32(frameToDecide)
		if round <= 0 {
			// Root cannot vote on any rounds from now on
			continue
		} else if round == 1 {
			// DBG(fmt.Sprintf("Event %c%d is DIRECTLY VOTING:\n", 'a'+rune(validatorId), frame))
			el.yesVote(frameToDecide, root)
		} else {
			// DBG(fmt.Sprintf("Event %c%d is AGGREGATING:\n", 'a'+rune(validatorId), frame))
			el.aggregateVotes(frameToDecide, frame, root)
		}
		// check if election is decided
		atropos, _ := el.chooseAtropos(frameToDecide)
		if atropos != nil {
			decidedAtropoi = append(decidedAtropoi, atropos)
		}
	}
	for _, atroposDecision := range decidedAtropoi {
		el.decidedFrameCleanup(atroposDecision.Frame)
	}
	return decidedAtropoi, nil
}

func (el *Election) newFrameToDecide(frame idx.Frame) {
	el.vote[frame] = make(map[hash.Event]map[idx.ValidatorID]bool)
	el.decided[frame] = make(map[idx.ValidatorID]bool)
	el.eventMap[frame] = make(map[idx.ValidatorID]hash.Event)
}

func (el *Election) getVoteVector(frame idx.Frame, event hash.Event) map[idx.ValidatorID]bool {
	if _, ok := el.vote[frame][event]; !ok {
		el.vote[frame][event] = make(map[idx.ValidatorID]bool)
	}
	return el.vote[frame][event]
}

func (el *Election) decidedFrameCleanup(frame idx.Frame) {
	// delete(el.vote, frame)
	// delete(el.decided, frame)
	// delete(el.eventMap, frame)
}

func (el *Election) yesVote(frame idx.Frame, root hash.Event) {
	observedRoots := el.observedRoots(root, frame)
	for _, candidateRoot := range observedRoots {
		el.eventMap[frame][candidateRoot.ValidatorID] = candidateRoot.EventID
		voteVector := el.getVoteVector(frame, root)
		voteVector[candidateRoot.ValidatorID] = true
		// DBG(fmt.Sprintf("For %c%d.\n", 'a'+rune(candidateRoot.ValidatorID), frame))
	}
	// DBG("\n")
}

func (el *Election) aggregateVotes(
	frameToDecide idx.Frame,
	frame idx.Frame,
	voterRoot hash.Event,
) error {
	observedRoots := el.observedRoots(voterRoot, frame-1)
	for _, validator := range el.validators.IDs() {
		if _, ok := el.decided[frameToDecide][validator]; ok {
			continue
		}
		yesVotes := el.validators.NewCounter()
		noVotes := el.validators.NewCounter()
		for _, observedRoot := range observedRoots {
			vote, ok := el.vote[frameToDecide][observedRoot.EventID][validator]
			if ok && vote {
				yesVotes.Count(observedRoot.ValidatorID)
				// DBG(fmt.Sprintf("For %c%d through %c%d, stake: %d.\n", 'a'+rune(validator), frameToDecide, 'a'+rune(observedRoot.ValidatorID), frame-1, el.validators.Get(observedRoot.ValidatorID)))
			} else {
				noVotes.Count(observedRoot.ValidatorID)
			}
		}
		// DBG(fmt.Sprintf("Total for %c%d: %d.\n", 'a'+rune(validator), frameToDecide, yesVotes.Sum()))
		if yesVotes.HasQuorum() || noVotes.HasQuorum() {
			if yesVotes.HasQuorum() {
				DBG("Decided Yes!\n")
			}
			el.decided[frameToDecide][validator] = yesVotes.HasQuorum()
		} else {
			voteVector := el.getVoteVector(frameToDecide, voterRoot)
			voteVector[validator] = yesVotes.Sum() >= noVotes.Sum()
		}
		// DBG("\n")
	}
	return nil
}

func (el *Election) chooseAtropos(frame idx.Frame) (*AtroposDecision, error) {
	if _, ok := el.atropos[frame]; ok {
		return nil, nil
	}
	for _, validatorId := range el.validators.SortedIDs() {
		decision, ok := el.decided[frame][validatorId]
		if !ok {
			return nil, nil // no new decisions
		}
		if decision {
			el.atropos[frame] = struct{}{}
			return &AtroposDecision{
				frame,
				el.eventMap[frame][validatorId],
			}, nil

			// return &el.eventMap[frame][validatorId], nil
		}
	}
	return nil, errors.New("all the roots are decided as 'no', which is possible only if more than 1/3W are Byzantine")
}

func (el *Election) observedRoots(root hash.Event, frame idx.Frame) []EventDescriptor {
	observedRoots := make([]EventDescriptor, 0, el.validators.Len())
	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.forklessCauses(root, frameRoot.EventID) {
			observedRoots = append(observedRoots, frameRoot)
		}
	}
	return observedRoots
}

func DBG(s string) {
	file, _ := os.OpenFile("DBG.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	file.WriteString(s)
	file.Close()
}
