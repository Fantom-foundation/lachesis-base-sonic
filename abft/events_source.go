package abft

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// EventSource is a callback for getting events from an external storage.
type EventSource interface {
	HasEvent(ltypes.EventHash) bool
	GetEvent(ltypes.EventHash) ltypes.Event
}
