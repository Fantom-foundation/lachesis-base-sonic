package electionv1

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/dag/tdag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/utils"
)

type fakeEdge struct {
	from hash.Event
	to   hash.Event
}

type (
	weights map[string]pos.Weight
)

type testExpected struct {
	DecidedFrame   idx.Frame
	DecidedAtropos string
	DecisiveRoots  map[string]bool
}

func TestEra(t *testing.T) {
	testProcessRoot(t,
		&testExpected{
			DecidedFrame:   0,
			DecidedAtropos: "d0_0",
			DecisiveRoots:  map[string]bool{"a2_2": true},
		},
		weights{
			"nodeA": 1,
			"nodeB": 1,
			"nodeC": 1,
			"nodeD": 1,
		}, `
a0_0  b0_0  c0_0  d0_0
║     ║     ║     ║
a1_1══╬═════╣     ║
║     ║     ║     ║
║     b1_1══╬═════╣
║     ║     ║     ║
║     ║╚════c1_1══╣
║     ║     ║     ║
║     ║╚═══─╫╩════d1_1
║     ║     ║     ║
a2_2══╬═════╬═════╣
║     ║     ║     ║
`)
}

func TestProcessRoot(t *testing.T) {

	t.Run("4 equalWeights notDecided", func(t *testing.T) {
		testProcessRoot(t,
			nil,
			weights{
				"nodeA": 1,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
			}, `
	a0_0  b0_0  c0_0  d0_0
	║     ║     ║     ║
	a1_1══╬═════╣     ║
	║     ║     ║     ║
	║╚════b1_1══╣     ║
	║     ║     ║     ║
	║     ║╚════c1_1══╣
	║     ║     ║     ║
	║     ║╚═══─╫╩════d1_1
	║     ║     ║     ║
	a2_2══╬═════╬═════╣
	║     ║     ║     ║
	`)
	})

	t.Run("4 equalWeights", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:   0,
				DecidedAtropos: "d0_0",
				DecisiveRoots:  map[string]bool{"a2_2": true},
			},
			weights{
				"nodeA": 1,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
			}, `
		a0_0  b0_0  c0_0  d0_0
		║     ║     ║     ║
		a1_1══╬═════╣     ║
		║     ║     ║     ║
		║     b1_1══╬═════╣
		║     ║     ║     ║
		║     ║╚════c1_1══╣
		║     ║     ║     ║
		║     ║╚═══─╫╩════d1_1
		║     ║     ║     ║
		a2_2══╬═════╬═════╣
		║     ║     ║     ║
		`)
	})

	t.Run("4 equalWeights missingRoot", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:   0,
				DecidedAtropos: "a0_0",
				DecisiveRoots:  map[string]bool{"a2_2": true},
			},
			weights{
				"nodeA": 1,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
			}, `
		a0_0  b0_0  c0_0  d0_0
		║     ║     ║     ║
		a1_1══╬═════╣     ║
		║     ║     ║     ║
		║╚════b1_1══╣     ║
		║     ║     ║     ║
		║╚═══─╫╩════c1_1  ║
		║     ║     ║     ║
		a2_2══╬═════╣     ║
		║     ║     ║     ║
		`)
	})

	t.Run("4 differentWeights", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:   0,
				DecidedAtropos: "a0_0",
				DecisiveRoots:  map[string]bool{"b2_2": true},
			},
			weights{
				"nodeA": math.MaxUint32/2 - 3,
				"nodeB": 1,
				"nodeC": 1,
				"nodeD": 1,
			}, `
		a0_0  b0_0  c0_0  d0_0
		║     ║     ║     ║
		a1_1══╬═════╣     ║
		║     ║     ║     ║
		║╚════+b1_1 ║     ║
		║     ║     ║     ║
		║╚═══─╫─════+c1_1 ║
		║     ║     ║     ║
		║╚═══─╫╩═══─╫╩════d1_1
		║     ║     ║     ║
		╠═════b2_2══╬═════╣
		║     ║     ║     ║
		`)
	})

	t.Run("4 differentWeights 4rounds", func(t *testing.T) {
		testProcessRoot(t,
			&testExpected{
				DecidedFrame:   0,
				DecidedAtropos: "a0_0",
				DecisiveRoots:  map[string]bool{"c2_2": true, "b2_2": true},
			},
			weights{
				"nodeA": 4,
				"nodeB": 2,
				"nodeC": 1,
				"nodeD": 1,
			}, `
	a0_0  b0_0  c0_0  d0_0
	║     ║     ║     ║
	a1_1══╣     ║     ║
	║     ║     ║     ║
	║     +b1_1═╬═════╣
	║     ║     ║     ║
	║╚═══─╫─════c1_1══╣
	║     ║     ║     ║
	║╚═══─╫─═══─╫╩════d1_1
	║     ║     ║     ║
	a2_2  ╣     ║     ║
	║     ║     ║     ║
	║╚════b2_2══╬═════╣
	║     ║     ║     ║
	║╚═══─╫╩════c2_2══╣
	║     ║     ║     ║
	║╚═══─╫╩═══─╫─════+d2_2
	`)
	})

}

type slot struct {
	frame       idx.Frame
	validatorID idx.ValidatorID
}

func testProcessRoot(
	t *testing.T,
	expected *testExpected,
	weights weights,
	dagAscii string,
) {
	t.Helper()
	assertar := assert.New(t)

	// events:
	ordered := make(tdag.TestEvents, 0)
	events := make(map[hash.Event]*tdag.TestEvent)
	frameRoots := make(map[idx.Frame][]EventDescriptor)
	vertices := make(map[hash.Event]slot)
	edges := make(map[fakeEdge]bool)

	nodes, _, _ := tdag.ASCIIschemeForEach(dagAscii, tdag.ForEachEvent{
		Process: func(_root dag.Event, name string) {
			root := _root.(*tdag.TestEvent)
			// store all the events
			ordered = append(ordered, root)

			events[root.ID()] = root

			slot := slot{
				frame:       frameOf(name),
				validatorID: root.Creator(),
			}
			vertices[root.ID()] = slot

			hsh := root.ID()
			frameRoots[frameOf(name)] = append(
				frameRoots[frameOf(name)],
				EventDescriptor{
					EventID:     &hsh,
					ValidatorID: slot.validatorID,
				},
			)

			// build edges to be able to fake forkless cause fn
			noPrev := false
			if strings.HasPrefix(name, "+") {
				noPrev = true
			}
			from := root.ID()
			for _, observed := range root.Parents() {
				if root.IsSelfParent(observed) && noPrev {
					continue
				}
				to := observed
				edge := fakeEdge{
					from: from,
					to:   to,
				}
				edges[edge] = true
			}
		},
	})

	validatorsBuilder := pos.NewBuilder()
	for _, node := range nodes {
		validatorsBuilder.Set(node, weights[utils.NameOf(node)])
	}
	validators := validatorsBuilder.Build()

	forklessCauseFn := func(a hash.Event, b hash.Event) bool {
		edge := fakeEdge{
			from: a,
			to:   b,
		}
		return edges[edge]
	}
	getFrameRootsFn := func(f idx.Frame) []EventDescriptor {
		return frameRoots[f]
	}

	// re-order events randomly, preserving parents order
	unordered := make(tdag.TestEvents, len(ordered))
	for i, j := range rand.Perm(len(ordered)) {
		unordered[i] = ordered[j]
	}
	ordered = unordered.ByParents()

	el := New(validators, forklessCauseFn, getFrameRootsFn)

	// processing:
	var alreadyDecided bool
	for _, root := range ordered {
		rootHash := root.ID()
		rootSlot, ok := vertices[rootHash]
		if !ok {
			t.Fatal("inconsistent vertices")
		}
		atropoi, err := el.ProcessRoot(rootSlot.frame, rootSlot.validatorID, rootHash)
		if err != nil {
			t.Fatal(err)
		}

		// checking:
		decisive := expected != nil && expected.DecisiveRoots[root.ID().String()]
		if decisive || alreadyDecided {
			assertar.NotNil(atropoi)
			assertar.Equal(expected.DecidedFrame, atropoi[0].Frame)
			assertar.Equal(expected.DecidedAtropos, atropoi[0].AtroposID.String())
			alreadyDecided = true
		} else {
			assertar.Equal([]AtroposDecision{}, atropoi)
		}
	}
}

func frameOf(dsc string) idx.Frame {
	s := strings.Split(dsc, "_")[1]
	h, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return idx.Frame(h)
}