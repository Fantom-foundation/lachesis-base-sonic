package electionv1

import (
	"container/heap"
	"errors"
	"os"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/viterin/vek/vek32"
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

	votes        map[hash.Event][]float32
	stake        []float32
	valMap       map[idx.ValidatorID]idx.Validator
	valNum       idx.Validator
	maxSeenFrame idx.Frame

	deliveryBuffer heapBuffer
	frameToDeliver idx.Frame
}

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
	el.maxSeenFrame = 0
	el.validators = validators
	el.votes = make(map[hash.Event][]float32)
	el.stake = make([]float32, 0, validators.Len())
	el.valNum = validators.Len()
	el.valMap = validators.Idxs()
}

func (el *Election) newFrameToDecide(frame idx.Frame) {
	el.decided[frame] = make(map[idx.ValidatorID]bool)
	el.eventMap[frame] = make(map[idx.ValidatorID]hash.Event)
}

func (el *Election) decidedFrameCleanup(frame idx.Frame) {
	delete(el.decided, frame)
	delete(el.eventMap, frame)
}

func (el *Election) PutInYourVotes(voterMatr []float32, frame idx.Frame, observedRoots []EventDescriptor) {
	// frame=2: voteMatr = [] -1 -1 -1 -1 -1]
	// frame=3: voteMatr = [agg1, agg2, agg3, agg4, agg5] -1 -1 -1 -1 -1]
	for idx := idx.Validator(0); idx < el.valNum; idx++ {
		voterMatr = append(voterMatr, -1.)
	}

	// frame=2: voteMatr = [seen1, seen2, seen3, seen4, seen5]
	// frame=3: voteMatr = [agg1, agg2, agg3, agg4, agg5, seen1, seen2, seen3, seen4, seen5]
	// All should be in range [-1, 1]
	for _, seenRoot := range observedRoots {
		voterMatr[idx.Validator(frame-2)*el.valNum+el.valMap[seenRoot.ValidatorID]] = 1.
	}
}

// ProcessRoot calculates Atropos votes only for the new root.
// If this root observes that the current election is decided, then return decided Atropoi
func (el *Election) ProcessRoot(
	frame idx.Frame,
	validatorId idx.ValidatorID,
	voterRoot hash.Event,
) ([]*AtroposDecision, error) {
	el.maxSeenFrame = max(el.maxSeenFrame, frame)
	if _, ok := el.eventMap[frame]; !ok {
		el.newFrameToDecide(frame)
	}
	el.eventMap[frame][validatorId] = voterRoot

	if frame == 1 {
		return []*AtroposDecision{}, nil
	} else if frame == 2 {
		el.votes[voterRoot] = make([]float32, 0, el.valNum)
		el.PutInYourVotes(el.votes[voterRoot], frame, el.observedRoots(voterRoot, frame-1))
		return []*AtroposDecision{}, nil
	}
	// valNum=5
	// frame=2: voteMatr = [] 0 0 0 0 0]
	// frame=3: voteMatr = [. . . . .] 0 0 0 0 0]
	voterMatr := make([]float32, (frame-2)*idx.Frame(el.valNum), (frame-1)*idx.Frame(el.valNum))
	decisionMatr := make([]float32, len(voterMatr))

	observedRoots := el.observedRoots(voterRoot, frame-1)
	stakeAccul := float32(0)
	for _, seenRoot := range observedRoots {
		vek32.Add_Inplace(voterMatr, el.votes[seenRoot.EventID])
		stakeAccul += float32(el.validators.GetWeightByIdx(el.validators.GetIdx(seenRoot.ValidatorID)))
	}
	copy(decisionMatr, voterMatr)
	vek32.Div_Inplace(voterMatr, vek32.Abs(voterMatr))

	el.PutInYourVotes(voterMatr, frame, observedRoots)
	vek32.MulNumber_Inplace(voterMatr, float32(el.validators.GetWeightByIdx(el.validators.GetIdx(validatorId))))

	Q := (4.*float32(el.validators.TotalWeight()) - 3*stakeAccul) / 4
	yesDecisions := vek32.GtNumber(decisionMatr, Q)
	noDecisions := vek32.LtNumber(decisionMatr, -Q)

	for f := el.frameToDeliver; f <= el.maxSeenFrame-2; f++ {
		for _, v := range el.validators.SortedIDs() {
			offset := (idx.Validator(f)-1)*el.valNum + el.validators.GetIdx(v)
			if yesDecisions[offset] {
				heap.Push(&el.deliveryBuffer, AtroposDecision{f, el.eventMap[f][v]})
				el.decidedFrameCleanup(f)
				break
			}
			if !noDecisions[offset] {
				break
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
// together and allows for efficient delivery of whole sequence when min frame Atropos of sequence aligns with 'frameToDeliver'
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
