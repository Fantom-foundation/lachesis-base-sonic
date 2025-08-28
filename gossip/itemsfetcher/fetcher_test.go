package itemsfetcher_test

import (
	"github.com/Fantom-foundation/lachesis-base/gossip/itemsfetcher"
	"testing"
	"time"
)

func TestFetcher(t *testing.T) {

	fetcher := itemsfetcher.New(itemsfetcher.Config{
		ForgetTimeout:       1 * time.Minute,
		ArriveTimeout:       1000 * time.Millisecond,
		GatherSlack:         100 * time.Millisecond,
		HashLimit:           10000,
		MaxBatch:            2,
		MaxParallelRequests: 1,
		MaxQueuedBatches:    2,
	}, itemsfetcher.Callback{
		OnlyInterested: func(ids []interface{}) []interface{} {
			return ids // we are interested in any announced item
		},
		Suspend: func() bool {
			return false
		},
	})
	fetcher.Start()
	defer fetcher.Stop()

	announcedIds1 := []interface{}{"eventA", "eventB", "eventC"}
	announcedIds2 := []interface{}{"eventD", "eventE"}
	fetchedIds := make([]interface{}, 0, 5)

	fetchItemsFn := func(ids []interface{}) error {
		fetchedIds = append(fetchedIds, ids...)
		return nil
	}

	err := fetcher.NotifyAnnounces("peer1", announcedIds1, time.Now(), fetchItemsFn)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Millisecond)
	if len(fetchedIds) != 3 {
		t.Errorf("unexpected fetchedIds: %v", fetchedIds)
	}

	err = fetcher.NotifyAnnounces("peer1", announcedIds2, time.Now(), fetchItemsFn)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Millisecond)
	if len(fetchedIds) != 5 {
		t.Errorf("unexpected fetchedIds: %v", fetchedIds)
	}
}
