package abft

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// EventSource is a callback for getting events from an external storage.
type EventSource interface {
	HasEvent(hash.EventHash) bool
	GetEvent(hash.EventHash) ltypes.Event
}
