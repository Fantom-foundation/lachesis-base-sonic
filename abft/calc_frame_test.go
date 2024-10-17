package abft

import (
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag/tdag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
)

func TestCalFrameIdx_10000(t *testing.T) {
	testCalcFrameIdx(t, 10000)
}

// testCalcFrameIdx verifies that lagging validator calculates correct frame numbers after a (large) pause
func testCalcFrameIdx(t *testing.T, gap int) {
	nodes := tdag.GenNodes(2)
	// Give one validator quorum power to advance the frames on it's own
	lch, _, store, _ := FakeLachesis(nodes, []pos.Weight{1, 3})

	laggyGenesis := processTestEvent(t, lch, store, nodes[0], 1, hash.Events{})
	parentEvent := processTestEvent(t, lch, store, nodes[1], 1, hash.Events{})
	for i := 0; i < gap; i++ {
		parentEvent = processTestEvent(t, lch, store, nodes[1], idx.Event(parentEvent.Seq()+1), hash.Events{parentEvent.ID()})
	}
	// Lagging validator creates an event after a frame gap
	finalEvent := processTestEvent(t, lch, store, nodes[0], laggyGenesis.Seq()+1, hash.Events{laggyGenesis.ID(), parentEvent.ID()})

	if want, got := laggyGenesis.Frame()+idx.Frame(gap)+1, finalEvent.Frame(); want != got {
		t.Errorf("expected calculated frame number of lagging validator to be: %d, got: %d", gap, finalEvent.Frame())
	}
}

var maxLamport idx.Lamport = 0

// processTestEvent builds and pipes the event through main Lacehsis' DAG manipulation pipeline
func processTestEvent(t *testing.T, lch *TestLachesis, store *EventStore, validatorId idx.ValidatorID, seq idx.Event, parents hash.Events) *tdag.TestEvent {
	event := &tdag.TestEvent{}
	event.SetSeq(seq)
	event.SetCreator(validatorId)
	event.SetParents(parents)
	maxLamport = maxLamport + 1
	event.SetLamport(maxLamport)
	event.SetEpoch(lch.store.GetEpoch())
	if err := lch.Build(event); err != nil {
		t.Errorf("error while building event for validator: %d, seq: %d, err: %v", validatorId, seq, err)
	}
	// default sample hash assigned through Build is enough
	store.SetEvent(event)
	if err := lch.Process(event); err != nil {
		t.Errorf("error while processing event for validator: %d, seq: %d, err: %v", validatorId, seq, err)
	}
	return event
}
