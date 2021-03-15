package dagprocessor

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/dag/tdag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
)

func TestProcessor(t *testing.T) {
	for try := int64(0); try < 500; try++ {
		testProcessor(t, try)
	}
}

var maxGroupSize = dag.Metric{
	Num:  50,
	Size: 50 * 50,
}

func shuffleIntoChunks(inEvents dag.Events) []dag.Events {
	if len(inEvents) == 0 {
		return nil
	}
	var chunks []dag.Events
	var lastChunk dag.Events
	var lastChunkSize dag.Metric
	for _, rnd := range rand.Perm(len(inEvents)) {
		e := inEvents[rnd]
		if rand.Intn(10) == 0 || lastChunkSize.Num+1 >= maxGroupSize.Num || lastChunkSize.Size+uint64(e.Size()) >= maxGroupSize.Size {
			chunks = append(chunks, lastChunk)
			lastChunk = dag.Events{}
		}
		lastChunk = append(lastChunk, e)
		lastChunkSize.Num++
		lastChunkSize.Size += uint64(e.Size())
	}
	chunks = append(chunks, lastChunk)
	return chunks
}

func testProcessor(t *testing.T, try int64) {
	nodes := tdag.GenNodes(5)

	var ordered dag.Events
	_ = tdag.ForEachRandEvent(nodes, 10, 3, nil, tdag.ForEachEvent{
		Process: func(e dag.Event, name string) {
			ordered = append(ordered, e)
		},
		Build: func(e dag.MutableEvent, name string) error {
			e.SetEpoch(1)
			e.SetFrame(idx.Frame(e.Seq()))
			return nil
		},
	})

	limit := dag.Metric{
		Num:  idx.Event(len(ordered)),
		Size: uint64(ordered.Metric().Size),
	}
	semaphore := datasemaphore.New(limit, func(received dag.Metric, processing dag.Metric, releasing dag.Metric) {
		t.Fatal("events semaphore inconsistency")
	})
	config := DefaultConfig()
	config.EventsBufferLimit = limit

	checked := 0

	highestLamport := idx.Lamport(0)
	processed := make(map[hash.Event]dag.Event)
	mu := sync.RWMutex{}
	processor := New(semaphore, config, Callback{
		Event: EventCallback{
			Process: func(e dag.Event) error {
				mu.Lock()
				defer mu.Unlock()
				if _, ok := processed[e.ID()]; ok {
					t.Fatalf("%s already processed", e.String())
					return nil
				}
				for _, p := range e.Parents() {
					if _, ok := processed[p]; !ok {
						t.Fatalf("got %s before parent %s", e.String(), p.String())
						return nil
					}
				}
				if highestLamport < e.Lamport() {
					highestLamport = e.Lamport()
				}
				processed[e.ID()] = e
				return nil
			},

			Released: func(e dag.Event, peer string, err error) {
				if err != nil {
					t.Fatalf("%s unexpectedly dropped with '%s'", e.String(), err)
				}
			},

			Exists: func(e hash.Event) bool {
				mu.RLock()
				defer mu.RUnlock()
				return processed[e] != nil
			},

			Get: func(id hash.Event) dag.Event {
				mu.RLock()
				defer mu.RUnlock()
				return processed[id]
			},

			OnlyInterested: func(ids hash.Events) hash.Events {
				mu.RLock()
				defer mu.RUnlock()
				onlyInterested := make(hash.Events, 0, len(ids))
				for _, id := range ids {
					if processed[id] != nil {
						onlyInterested = append(onlyInterested, id)
					}
				}
				return onlyInterested
			},
			CheckParents: func(e dag.Event, parents dag.Events) error {
				mu.RLock()
				defer mu.RUnlock()
				checked++
				if e.Frame() != idx.Frame(e.Seq()) {
					return errors.New("malformed event frame")
				}
				return nil
			},
			CheckParentless: func(inEvents dag.Events, checked func(ee dag.Events, errs []error)) {
				chunks := shuffleIntoChunks(inEvents)
				for _, chunk := range chunks {
					checked(chunk, make([]error, len(chunk)))
				}
			},
		},
		PeerMisbehaviour: func(peer string, err error) bool {
			return rand.Intn(2) == 0
		},
		HighestLamport: func() idx.Lamport {
			return highestLamport
		},
	})
	// shuffle events
	chunks := shuffleIntoChunks(ordered)

	// process events
	processor.Start()
	wg := sync.WaitGroup{}
	for _, chunk := range chunks {
		wg.Add(1)
		err := processor.Enqueue("", chunk, rand.Intn(2) == 0, func(events hash.Events) {}, func() {
			wg.Done()
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	processor.Stop()

	// everything is processed
	for _, e := range ordered {
		if _, ok := processed[e.ID()]; !ok {
			t.Fatal("event wasn't processed")
		}
	}
	if checked != len(processed) {
		t.Fatal("not all the events were checked")
	}
}

func TestProcessorReleasing(t *testing.T) {
	for try := int64(0); try < 100; try++ {
		testProcessorReleasing(t, 200, try)
	}
}

func testProcessorReleasing(t *testing.T, maxEvents int, try int64) {
	nodes := tdag.GenNodes(5)

	var ordered dag.Events
	r := rand.New(rand.NewSource(try))
	_ = tdag.ForEachRandEvent(nodes, 10, 3, r, tdag.ForEachEvent{
		Process: func(e dag.Event, name string) {
			ordered = append(ordered, e)
		},
		Build: func(e dag.MutableEvent, name string) error {
			e.SetEpoch(1)
			e.SetFrame(idx.Frame(e.Seq()))
			return nil
		},
	})

	limit := dag.Metric{
		Num:  idx.Event(r.Intn(maxEvents)),
		Size: uint64(r.Intn(maxEvents * 100)),
	}
	limitPlus1group := dag.Metric{
		Num:  limit.Num + maxGroupSize.Num,
		Size: limit.Size + maxGroupSize.Size,
	}
	semaphore := datasemaphore.New(limitPlus1group, func(received dag.Metric, processing dag.Metric, releasing dag.Metric) {
		t.Fatal("events semaphore inconsistency")
	})
	config := DefaultConfig()
	config.EventsBufferLimit = limit

	released := uint32(0)

	highestLamport := idx.Lamport(0)
	processed := make(map[hash.Event]dag.Event)
	mu := sync.RWMutex{}
	processor := New(semaphore, config, Callback{
		Event: EventCallback{
			Process: func(e dag.Event) error {
				mu.Lock()
				defer mu.Unlock()
				if _, ok := processed[e.ID()]; ok {
					t.Fatalf("%s already processed", e.String())
					return nil
				}
				for _, p := range e.Parents() {
					if _, ok := processed[p]; !ok {
						t.Fatalf("got %s before parent %s", e.String(), p.String())
						return nil
					}
				}
				if r.Intn(10) == 0 {
					return errors.New("testing error")
				}
				if r.Intn(10) == 0 {
					time.Sleep(time.Microsecond * 100)
				}
				if highestLamport < e.Lamport() {
					highestLamport = e.Lamport()
				}
				processed[e.ID()] = e
				return nil
			},

			Released: func(e dag.Event, peer string, err error) {
				mu.Lock()
				defer mu.Unlock()
				atomic.AddUint32(&released, 1)
			},

			Exists: func(e hash.Event) bool {
				mu.RLock()
				defer mu.RUnlock()
				return processed[e] != nil
			},

			Get: func(id hash.Event) dag.Event {
				mu.RLock()
				defer mu.RUnlock()
				return processed[id]
			},

			OnlyInterested: func(ids hash.Events) hash.Events {
				mu.RLock()
				defer mu.RUnlock()
				onlyInterested := make(hash.Events, 0, len(ids))
				for _, id := range ids {
					if processed[id] != nil {
						onlyInterested = append(onlyInterested, id)
					}
				}
				return onlyInterested
			},
			CheckParents: func(e dag.Event, parents dag.Events) error {
				if r.Intn(10) == 0 {
					return errors.New("testing error")
				}
				if r.Intn(10) == 0 {
					time.Sleep(time.Microsecond * 100)
				}
				return nil
			},
			CheckParentless: func(inEvents dag.Events, checked func(ee dag.Events, errs []error)) {
				chunks := shuffleIntoChunks(inEvents)
				for _, chunk := range chunks {
					errs := make([]error, len(chunk))
					for i := range errs {
						if r.Intn(10) == 0 {
							errs[i] = errors.New("testing error")
						}
					}
					if r.Intn(10) == 0 {
						time.Sleep(time.Microsecond * 100)
					}
					checked(chunk, errs)
				}
			},
		},
		PeerMisbehaviour: func(peer string, err error) bool {
			return r.Intn(2) == 0
		},
		HighestLamport: func() idx.Lamport {
			return highestLamport
		},
	})
	// duplicate some events
	ordered = append(ordered, ordered[:r.Intn(len(ordered))]...)
	// shuffle events
	chunks := shuffleIntoChunks(ordered)

	// process events
	processor.Start()
	wg := sync.WaitGroup{}
	for _, chunk := range chunks {
		wg.Add(1)
		err := processor.Enqueue("", chunk, r.Intn(2) == 0, func(events hash.Events) {}, func() {
			wg.Done()
		})
		if err != nil {
			t.Fatal(err)
		}
		if r.Intn(10) == 0 {
			processor.Clear()
		}
	}
	wg.Wait()
	processor.Clear()
	if processor.eventsSemaphore.Processing().Num != 0 {
		t.Fatal("not all the events were released", processor.eventsSemaphore.Processing().Num)
	}
	processor.Stop()

	// everything is released
	if uint32(len(ordered)) != released {
		t.Fatal("not all the events were released", len(ordered), released)
	}
}
