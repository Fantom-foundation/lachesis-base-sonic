package lachesis

import (
	"github.com/Fantom-foundation/lachesis-base/types"
)

// Block is a part of an ordered chain of batches of events.
type Block struct {
	Atropos  types.Event
	Cheaters Cheaters
}
