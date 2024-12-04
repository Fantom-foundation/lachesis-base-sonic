package lachesis

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// Block is a part of an ordered chain of batches of events.
type Block struct {
	Atropos  ltypes.EventHash
	Cheaters Cheaters
}
