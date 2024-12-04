package types

import (
	"math/rand"

	"sync"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
)

// FakeHash generates random fake hash for testing purpose.
func FakeHash(seed ...int64) (h common.Hash) {
	randRead := rand.Read

	if len(seed) > 0 {
		src := rand.NewSource(seed[0])
		rnd := rand.New(src) // nolint:gosec
		randRead = rnd.Read
	}

	_, err := randRead(h[:])
	if err != nil {
		panic(err)
	}
	return
}

var (
	nodeNameDictMu  sync.RWMutex
	eventNameDictMu sync.RWMutex

	// nodeNameDict is an optional dictionary to make node address human readable in log.
	nodeNameDict = make(map[idx.ValidatorID]string)

	// eventNameDict is an optional dictionary to make events human readable in log.
	eventNameDict = make(map[Event]string)
)

// SetNodeName sets an optional human readable alias of node address in log.
func SetNodeName(n idx.ValidatorID, name string) {
	nodeNameDictMu.Lock()
	defer nodeNameDictMu.Unlock()

	nodeNameDict[n] = name
}

// SetEventName sets an optional human readable alias of event hash in log.
func SetEventName(e Event, name string) {
	eventNameDictMu.Lock()
	defer eventNameDictMu.Unlock()

	eventNameDict[e] = name
}

// GetNodeName gets an optional human readable alias of node address.
func GetNodeName(n idx.ValidatorID) string {
	nodeNameDictMu.RLock()
	defer nodeNameDictMu.RUnlock()

	return nodeNameDict[n]
}

// GetEventName gets an optional human readable alias of event hash.
func GetEventName(e Event) string {
	eventNameDictMu.RLock()
	defer eventNameDictMu.RUnlock()

	return eventNameDict[e]
}
