package electionv1

import (
	"container/heap"
	"os"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/viterin/vek"
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

	eventMap  map[idx.Frame]map[idx.ValidatorID]hash.Event
	delivered map[idx.Frame]struct{}

	votes           map[hash.Event][]float32
	valMap          map[idx.ValidatorID]idx.Validator
	valNum          idx.Frame
	emptyVoteVector []float32

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
	el.eventMap = make(map[idx.Frame]map[idx.ValidatorID]hash.Event)
	el.deliveryBuffer = make(heapBuffer, 0)
	heap.Init(&el.deliveryBuffer)
	el.frameToDeliver = 1
	el.validators = validators
	el.votes = make(map[hash.Event][]float32)
	el.valNum = idx.Frame(validators.Len())
	el.valMap = validators.Idxs()
	el.delivered = make(map[idx.Frame]struct{})
	el.emptyVoteVector = vek32.Repeat(-1., int(el.valNum))
}

func (el *Election) newRoot(frame idx.Frame, validatorId idx.ValidatorID, root hash.Event) {
	if _, ok := el.eventMap[frame]; !ok {
		el.eventMap[frame] = make(map[idx.ValidatorID]hash.Event)
	}
	el.eventMap[frame][validatorId] = root
}

func (el *Election) decidedFrameCleanup(frame idx.Frame) {
	delete(el.eventMap, frame)
}

// ProcessRoot calculates Atropos votes only for the new root.
// If this root observes that the current election is decided, then return decided Atropoi
func (el *Election) ProcessRoot(
	frame idx.Frame,
	validatorId idx.ValidatorID,
	voterRoot hash.Event,
) ([]*AtroposDecision, error) {
	vek.SetAcceleration(false)
	el.newRoot(frame, validatorId, voterRoot)
	if frame == 1 {
		el.votes[voterRoot] = make([]float32, 0)
		return []*AtroposDecision{}, nil
	}
	voterMatr := make([]float32, (frame-2)*el.valNum, (frame-1)*el.valNum)
	voteVec := vek32.Repeat(-1., int(el.valNum))

	observedRoots := el.observedRoots(voterRoot, frame-1)
	stakeAccul := float32(0)
	for _, seenRoot := range observedRoots {
		voteVec[el.valMap[seenRoot.ValidatorID]] = 1.
		stakeAccul += float32(el.validators.GetWeightByIdx(el.validators.GetIdx(seenRoot.ValidatorID)))

		vek32.Add_Inplace(voterMatr, el.votes[seenRoot.EventID])
	}
	if frame > 2 {
		el.decideRoots(frame, voterMatr, stakeAccul)
		vek32.Div_Inplace(voterMatr, vek32.Abs(voterMatr))
	}
	voterMatr = append(voterMatr, voteVec...)
	vek32.MulNumber_Inplace(voterMatr, float32(el.validators.GetWeightByIdx(el.valMap[validatorId])))
	el.votes[voterRoot] = voterMatr
	return el.alignedAtropoi(), nil
}

func (el *Election) decideRoots(frame idx.Frame, aggregationMatr []float32, seenRootsStake float32) {
	Q := (4.*float32(el.validators.TotalWeight()) - 3*seenRootsStake) / 4
	yesDecisions := vek32.GtNumber(aggregationMatr, Q)
	noDecisions := vek32.LtNumber(aggregationMatr, -Q)

	for f := range el.eventMap {
		if _, ok := el.delivered[f]; ok || f >= frame-1 {
			continue
		}
		for _, v := range el.validators.SortedIDs() {
			offset := (f-1)*el.valNum + idx.Frame(el.validators.GetIdx(v))
			if yesDecisions[offset] {
				heap.Push(&el.deliveryBuffer, &AtroposDecision{f, el.eventMap[f][v]})
				el.decidedFrameCleanup(f)
				break
			}
			if !noDecisions[offset] {
				break
			}
		}
	}

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
		el.delivered[el.deliveryBuffer[0].Frame] = struct{}{}
		deliveredAtropoi = append(deliveredAtropoi, heap.Pop(&el.deliveryBuffer).(*AtroposDecision))
		el.frameToDeliver++
	}
	for f := el.frameToDeliver - idx.Frame(len(deliveredAtropoi)); f < el.frameToDeliver; f++ {
		for v := range el.eventMap[f] {
			delete(el.votes, el.eventMap[f][v])
		}
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
