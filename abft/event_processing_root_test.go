package abft

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/dag/tdag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

func TestLachesisClassicRoots(t *testing.T) {
	testSpecialNamedRoots(t, `
A1.01  B1.01  C1.01  D1.01  // 1
║      ║      ║      ║
║      ╠──────╫───── d1.02
║      ║      ║      ║
║      b1.02 ─╫──────╣
║      ║      ║      ║
║      ╠──────╫───── d1.03
a1.02 ─╣      ║      ║
║      ║      ║      ║
║      b1.03 ─╣      ║
║      ║      ║      ║
║      ╠──────╫───── d1.04
║      ║      ║      ║
║      ╠───── c1.02  ║
║      ║      ║      ║
║      b1.04 ─╫──────╣
║      ║      ║      ║     // 2
╠──────╫──────╫───── D2.05
║      ║      ║      ║
A2.03 ─╫──────╫──────╣
║      ║      ║      ║
a2.04 ─╫──────╣      ║
║      ║      ║      ║
║      B2.05 ─╫──────╣
║      ║      ║      ║
║      ╠──────╫───── d2.06
a2.05 ─╣      ║      ║
║      ║      ║      ║
╠──────╫───── C2.03  ║
║      ║      ║      ║
╠──────╫──────╫───── d2.07
║      ║      ║      ║
╠───── b2.06  ║      ║
║      ║      ║      ║     // 3
║      B3.07 ─╫──────╣
║      ║      ║      ║
A3.06 ─╣      ║      ║
║      ╠──────╫───── D3.08
║      ║      ║      ║
║      ║      ╠───── d309
╠───── b3.08  ║      ║
║      ║      ║      ║
╠───── b3.09  ║      ║
║      ║      C3.04 ─╣
a3.07 ─╣      ║      ║
║      ║      ║      ║
║      b3.10 ─╫──────╣
║      ║      ║      ║
a3.08 ─╣      ║      ║
║      ╠──────╫───── d3.10
║      ║      ║      ║
╠───── b3.11  ║      ║     // 4
║      ║      ╠───── D4.11
║      ║      ║      ║
║      B4.12 ─╫──────╣
║      ║      ║      ║
`)
}

func TestLachesisRandomRoots(t *testing.T) {
	// generated by codegen4LachesisRandomRoot()
	testSpecialNamedRoots(t, `
 A1.01    
 ║         ║        
 ╠════════ B1.01    
 ║         ║         ║        
 ╠════════─╫─═══════ C1.01    
 ║         ║         ║         ║        
 ╠════════─╫─═══════─╫─═══════ D1.01    
 ║         ║         ║         ║        
 a1.02════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ║         b1.02════─╫─════════╣        
 ║         ║         ║         ║        
 ║         ║         c1.02═════╣        
 ║         ║         ║         ║        
 a1.03════─╫─════════╣         ║        
 ║         ║         ║         ║        
 ╠════════ B2.03     ║         ║        
 ║         ║║        ║         ║        
 ║         ║╚═══════─╫─═══════ d1.02    
 ║         ║         ║         ║        
 ║         ║         C2.03═════╣        
 ║         ║         ║         ║        
 A2.04════─╫─════════╣         ║        
 ║         ║         ║         ║        
 ║         b2.04═════╣         ║        
 ║         ║║        ║         ║        
 ║         ║╚═══════─╫─═══════ D2.03    
 ║         ║         ║         ║        
 ║         ║         c2.04═════╣        
 ║         ║         ║         ║        
 ║         ║         ╠════════ d2.04    
 ║         ║         ║         ║        
 A3.05════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ╠════════ B3.05     ║         ║        
 ║         ║         ║         ║        
 ║         ╠════════ C3.05     ║        
 ║         ║         ║         ║        
 ║         ╠════════─╫─═══════ D3.05    
 ║         ║         ║         ║        
 a3.06════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ║         b3.06════─╫─════════╣        
 ║         ║         ║         ║        
 ║         ║         c3.06═════╣        
 ║         ║         ║         ║        
 ║         B4.07═════╣         ║        
 ║         ║         ║         ║        
 ║         ║         ╠════════ d3.06    
 ║         ║         ║         ║        
 A4.07════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 a4.08═════╣         ║         ║        
 ║║        ║         ║         ║        
 ║╚═══════─╫─═══════ C4.07     ║        
 ║         ║         ║         ║        
 ║         b4.08═════╣         ║        
 ║         ║         ║         ║        
 a4.09═════╣         ║         ║        
 ║3        ║         ║         ║        
 ║╚═══════─╫─═══════─╫─═══════ D4.07    
 ║         ║         ║         ║        
 ║         ║         c4.08═════╣        
 ║         ║         ║         ║        
 ║         b4.09═════╣         ║        
 ║         ║         ║         ║        
 ║         ╠════════ c4.09     ║        
 ║         ║         ║         ║        
 A5.10════─╫─════════╣         ║        
 ║         ║         ║         ║        
 ╠════════ B5.10     ║         ║        
 ║         ║3        ║         ║        
 ║         ║╚═══════─╫─═══════ d4.08    
 ║║        ║         ║         ║        
 ║╚═══════─╫─═══════─╫─═══════ D5.09    
 ║         ║         ║         ║        
 ║         ║         C5.10═════╣        
 ║         ║         ║         ║        
 ╠════════─╫─═══════─╫─═══════ d5.10    
 ║         ║         ║         ║        
 a5.11════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ╠════════ b5.11     ║         ║        
 ║         ║         ║         ║        
 ║         ╠════════ c5.11     ║        
 ║         ║         ║         ║        
 A6.12════─╫─════════╣         ║        
 ║         ║         ║         ║        
 ║         ╠════════─╫─═══════ d5.11    
 ║         ║         ║         ║        
 ║         b5.12════─╫─════════╣        
 ║         ║         ║         ║        
 ║         ╠════════ C6.12     ║        
 ║         ║         ║         ║        
 ╠════════─╫─═══════─╫─═══════ D6.12    
 ║         ║         ║         ║        
 a6.13════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ║         B6.13════─╫─════════╣        
 ║         ║         ║         ║        
 a6.14═════╣         ║         ║        
 ║║        ║         ║         ║        
 ║╚═══════─╫─═══════ c6.13     ║        
 ║         ║         ║         ║        
 ╠════════─╫─═══════ C7.14     ║        
 ║║        ║         ║         ║        
 ║╚═══════─╫─═══════─╫─═══════ d6.13    
 ║         ║         ║         ║        
 ║         b6.14════─╫─════════╣        
 ║         ║         ║         ║        
 a6.15═════╣         ║         ║        
 ║         ║         ║         ║        
 ║         B7.15═════╣         ║        
 ║         ║║        ║         ║        
 ║         ║╚═══════─╫─═══════ d6.14    
 ║         ║         ║         ║        
 ║         ║         c7.15═════╣        
 ║         ║         ║         ║        
 ╠════════─╫─═══════─╫─═══════ D7.15    
 ║         ║         ║         ║        
 A7.16════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ║         b7.16════─╫─════════╣        
 ║         ║         ║         ║        
 ║         ║         c7.16═════╣        
 ║         ║         ║         ║        
 a7.17════─╫─════════╣         ║        
 ║         ║         ║         ║        
 ║         ║         ╠════════ d7.16    
 ║         ║         ║         ║        
 ║         b7.17════─╫─════════╣        
 ║         ║         ║         ║        
 ║         ║         c7.17═════╣        
 ║         ║         ║         ║        
 a7.18════─╫─════════╣         ║        
 ║         ║         ║         ║        
 ╠════════─╫─═══════ c7.18     ║        
 ║║        ║         ║         ║        
 ║╚═══════─╫─═══════─╫─═══════ d7.17    
 ║         ║         ║         ║        
 ║         B8.18════─╫─════════╣        
 ║         ║         ║         ║        
 ║         b8.19═════╣         ║        
 ║         ║║        ║         ║        
 ║         ║╚═══════─╫─═══════ D8.18    
 ║         ║         ║         ║        
 A8.19════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ╠════════─╫─═══════ C8.19     ║        
 ║         ║         ║         ║        
 ╠════════─╫─═══════─╫─═══════ d8.19    
 ║         ║         ║         ║        
 a8.20════─╫─═══════─╫─════════╣        
 ║         ║         ║         ║        
 ║         B9.20════─╫─════════╣        
 ║         ║         ║         ║        
 ║         ║         C9.20═════╣        
 ║         ║         ║         ║        
 ║         ║         ╠════════ D9.20   
`)
}

/*
 * Utils:
 */

// testSpecialNamedRoots is a general test of root selection.
// Event name means:
// - 1st letter uppercase - event should be root;
// - 2nd number - frame where event should be in;
// - "." - separator;
// - tail - makes name unique;
func testSpecialNamedRoots(t *testing.T, scheme string) {
	t.Helper()
	//logger.SetTestMode(t)
	assertar := assert.New(t)

	// decode is a event name parser
	decode := func(name string) (frameN idx.Frame, isRoot bool) {
		n, err := strconv.ParseUint(strings.Split(name, ".")[0][1:2], 10, 64)
		if err != nil {
			panic(err.Error() + ". Name event " + name + " properly: <UpperCaseForRoot><FrameN><Engine>")
		}
		frameN = idx.Frame(n)

		isRoot = name == strings.ToUpper(name)
		return
	}

	// get nodes only
	nodes, _, _ := tdag.ASCIIschemeToDAG(scheme)
	// init abft
	lch, _, input, _ := NewCoreLachesis(nodes, nil)

	// process events
	_, _, names := tdag.ASCIIschemeForEach(scheme, tdag.ForEachEvent{
		Process: func(e dag.Event, name string) {
			input.SetEvent(e)
			assertar.NoError(
				lch.Process(e))
		},
		Build: func(e dag.MutableEvent, name string) error {
			e.SetEpoch(lch.store.GetEpoch())
			return lch.Build(e)
		},
	})

	// check each
	for name, event := range names {
		mustBeFrame, mustBeRoot := decode(name)
		var selfParentFrame idx.Frame
		if event.SelfParent() != nil {
			selfParentFrame = input.GetEvent(*event.SelfParent()).Frame()
		}
		// check root
		if !assertar.Equal(mustBeRoot, event.Frame() != selfParentFrame, name+" is root") {
			break
		}
		// check frame
		if !assertar.Equal(mustBeFrame, event.Frame(), "frame of "+name) {
			break
		}
	}
}

/*
// codegen4LachesisRandomRoot is for test data generation.
func codegen4LachesisRandomRoot() {
	nodes, events := inter.GenEventsByNode(4, 20, 2, nil)

	p, _, input := FakeLachesis(nodes)
	// process events
	config := inter.Events{}
	for _, ee := range events {
		config = append(config, ee...)
		for _, e := range ee {
			input.SetEvent(e)
			p.PushEventSync(e.ID())
		}
	}

	// set event names
	for _, e := range config {
		frame := p.FrameOfEvent(e.ID())
		_, isRoot := frame.Roots[e.Creator][e.ID()]
		oldName := hash.GetEventName(e.ID())
		newName := fmt.Sprintf("%s%d.%02d", oldName[0:1], frame.Engine, e.Seq)
		if isRoot {
			newName = strings.ToUpper(newName[0:1]) + newName[1:]
		}
		hash.SetEventName(e.ID(), newName)
	}

	fmt.Println(inter.DAGtoASCIIscheme(config))
}
*/
