package abft

import (
	"github.com/Fantom-foundation/lachesis-base/types"
	"github.com/Fantom-foundation/lachesis-base/types"
)

// EventSource is a callback for getting events from an external storage.
type EventSource interface {
	HasEvent(hash.Event) bool
	GetEvent(hash.Event) dag.Event
}
