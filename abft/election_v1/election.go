package electionv1

import (
	"container/heap"
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

	deliveryBuffer heapBuffer
	frameToDeliver idx.Frame
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
	el.deliveryBuffer = make(heapBuffer, 0)
	heap.Init(&el.deliveryBuffer)
	el.frameToDeliver = 1
	el.validators = validators
}

func (el *Election) newFrameToDecide(frame idx.Frame) {
	el.vote[frame] = make(map[hash.Event]map[idx.ValidatorID]bool)
	el.decided[frame] = make(map[idx.ValidatorID]bool)
	el.eventMap[frame] = make(map[idx.ValidatorID]hash.Event)
}

func (el *Election) decidedFrameCleanup(frame idx.Frame) {
	delete(el.vote, frame)
	delete(el.decided, frame)
	delete(el.eventMap, frame)
}

// ProcessRoot calculates Atropos votes only for the new root.
// If this root observes that the current election is decided, then return decided Atropos
func (el *Election) ProcessRoot(
	frame idx.Frame,
	validatorId idx.ValidatorID,
	root hash.Event,
) ([]*AtroposDecision, error) {
	// Iterate over all undecided frames
	if _, ok := el.vote[frame]; !ok {
		el.newFrameToDecide(frame)
	}
	for frameToDecide := range el.vote {
		round := int32(frame) - int32(frameToDecide)
		if round <= 0 {
			// Root cannot vote
			continue
		} else if round == 1 {
			// DBG(fmt.Sprintf("Event %c%d is DIRECTLY VOTING:\n", 'a'+rune(validatorId), frame))
			el.yesVote(frameToDecide, root)
		} else {
			// DBG(fmt.Sprintf("Event %c%d is AGGREGATING:\n", 'a'+rune(validatorId), frame))
			el.aggregateVotes(frameToDecide, frame, root) // check if election is decided
			atropos, _ := el.chooseAtropos(frameToDecide)
			if atropos != nil {
				heap.Push(&el.deliveryBuffer, atropos)
				el.decidedFrameCleanup(frameToDecide)
			}
		}
	}
	return el.alignedAtropoi(), nil
}

func (el *Election) getVoteVector(frame idx.Frame, event hash.Event) map[idx.ValidatorID]bool {
	if _, ok := el.vote[frame][event]; !ok {
		el.vote[frame][event] = make(map[idx.ValidatorID]bool)
	}
	return el.vote[frame][event]
}

func (el *Election) yesVote(frame idx.Frame, root hash.Event) {
	observedRoots := el.observedRoots(root, frame)
	for _, candidateRoot := range observedRoots {
		el.eventMap[frame][candidateRoot.ValidatorID] = candidateRoot.EventID
		voteVector := el.getVoteVector(frame, root)
		voteVector[candidateRoot.ValidatorID] = true
	}
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
			} else {
				noVotes.Count(observedRoot.ValidatorID)
			}
		}
		if yesVotes.HasQuorum() || noVotes.HasQuorum() {
			el.decided[frameToDecide][validator] = yesVotes.HasQuorum()
		} else {
			voteVector := el.getVoteVector(frameToDecide, voterRoot)
			voteVector[validator] = yesVotes.Sum() >= noVotes.Sum()
		}
	}
	return nil
}

func (el *Election) chooseAtropos(frame idx.Frame) (*AtroposDecision, error) {
	for _, validatorId := range el.validators.SortedIDs() {
		decision, ok := el.decided[frame][validatorId]
		if !ok {
			return nil, nil // no new decisions
		}
		if decision {
			return &AtroposDecision{
				frame,
				el.eventMap[frame][validatorId],
			}, nil
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

// alignedAtropoi pops and returns only continuous sequence of decided atropoi
// that start with `frameToDeliver` frame number
// example 1: frameToDeliver = 100, heapBuffer = [100, 101, 102], deliveredAtropoi = [100, 101, 102]
// example 2: frameToDeliver = 100, heapBuffer = [101, 102], deliveredAtropoi = []
// example 3: frameToDeliver = 100, heapBuffer = [100, 101, 104, 105], deliveredAtropoi = [100, 101]
func (el *Election) alignedAtropoi() []*AtroposDecision {
	deliveredAtropoi := make([]*AtroposDecision, 0)
	for len(el.deliveryBuffer) > 0 && el.deliveryBuffer[0].Frame == el.frameToDeliver {
		deliveredAtropoi = append(deliveredAtropoi, heap.Pop(&el.deliveryBuffer).(*AtroposDecision))
		el.frameToDeliver++
	}
	return deliveredAtropoi
}

// heapBuffer is a min-heap of Atropos decisions ordered by Frames.
// it is an easy to maintain structure that keeps continuous sequences (possibly multiple patches of them)
// together and allows for efficient delivery of the sequence when the minimal Atropos in the sequence aligns with 'frameToDeliver'
type heapBuffer []*AtroposDecision

func (h heapBuffer) Len() int           { return len(h) }
func (h heapBuffer) Less(i, j int) bool { return h[i].Frame < h[j].Frame }
func (h heapBuffer) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *heapBuffer) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*AtroposDecision))
}

func (h *heapBuffer) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
