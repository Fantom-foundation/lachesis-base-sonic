package epochcheck

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

var (
	// ErrNotRelevant indicates the event's epoch isn't equal to current epoch.
	ErrNotRelevant = errors.New("event is too old or too new")
	// ErrAuth indicates that event's creator isn't authorized to create events in current epoch.
	ErrAuth = errors.New("event creator isn't a validator")
)

// Reader returns currents epoch and its validators group.
type Reader interface {
	GetEpochValidators() (*ltypes.Validators, idx.EpochID)
}

// Checker which require only current epoch info
type Checker struct {
	reader Reader
}

func New(reader Reader) *Checker {
	return &Checker{
		reader: reader,
	}
}

// Validate event
func (v *Checker) Validate(e ltypes.Event) error {
	// check epoch first, because validators group is returned only for the current epoch
	validators, epoch := v.reader.GetEpochValidators()
	if e.Epoch() != epoch {
		return ErrNotRelevant
	}
	if !validators.Exists(e.Creator()) {
		return ErrAuth
	}
	return nil
}
