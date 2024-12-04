package eventcheck

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/eventcheck/parentscheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/ltypes"
	"github.com/Fantom-foundation/lachesis-base/ltypes/tdag"
)

type testReader struct{}

func (tr *testReader) GetEpochValidators() (*ltypes.Validators, idx.EpochID) {
	vb := ltypes.NewBuilder()
	vb.Set(1, 1)
	return vb.Build(), 1
}

func TestBasicEventValidation(t *testing.T) {
	var tests = []struct {
		e       ltypes.Event
		wantErr error
	}{
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(1)
			e.SetLamport(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			return e
		}(), nil},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(0)
			e.SetLamport(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			return e
		}(), basiccheck.ErrNotInited},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			return e
		}(), basiccheck.ErrNoParents},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(math.MaxInt32 - 1)
			e.SetLamport(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			return e
		}(), basiccheck.ErrHugeValue},
	}

	for _, tt := range tests {
		basicCheck := basiccheck.New()
		assert.Equal(t, tt.wantErr, basicCheck.Validate(tt.e))
	}
}

func TestEpochEventValidation(t *testing.T) {
	var tests = []struct {
		e       ltypes.Event
		wantErr error
	}{
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetEpoch(1)
			e.SetCreator(1)
			return e
		}(), nil},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetEpoch(2)
			e.SetCreator(1)
			return e
		}(), epochcheck.ErrNotRelevant},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetEpoch(1)
			e.SetCreator(2)
			return e
		}(), epochcheck.ErrAuth},
	}

	for _, tt := range tests {
		tr := new(testReader)
		epochCheck := epochcheck.New(tr)
		assert.Equal(t, tt.wantErr, epochCheck.Validate(tt.e))
	}
}

func TestParentsEventValidation(t *testing.T) {
	var tests = []struct {
		e         ltypes.Event
		pe        ltypes.Events
		wantErr   error
		wantPanic bool
	}{
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(2)
			e.SetCreator(1)
			selfParent := &tdag.TestEvent{}
			selfParent.SetLamport(1)
			selfParent.SetID([24]byte{1})
			e.SetParents(hash.EventHashes{selfParent.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(1)
				e.SetLamport(1)
				e.SetCreator(1)
				e.SetID([24]byte{1})
				return ltypes.Events{e}
			}(),
			nil, false},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(2)
			e.SetCreator(1)
			selfParent := &tdag.TestEvent{}
			selfParent.SetLamport(1)
			selfParent.SetID([24]byte{2})
			e.SetParents(hash.EventHashes{selfParent.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(1)
				e.SetLamport(1)
				e.SetCreator(1)
				e.SetID([24]byte{1})
				return ltypes.Events{e}
			}(),
			parentscheck.ErrWrongSelfParent, false},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(1)
			e.SetParents(hash.EventHashes{e.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(1)
				e.SetLamport(1)
				return ltypes.Events{e}
			}(),
			parentscheck.ErrWrongLamport, false},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(1)
			e.SetLamport(2)
			e.SetParents(hash.EventHashes{e.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(1)
				e.SetLamport(1)
				return ltypes.Events{e}
			}(),
			parentscheck.ErrWrongSelfParent, false},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(2)
			selfParent := &tdag.TestEvent{}
			selfParent.SetLamport(1)
			selfParent.SetID([24]byte{1})
			e.SetParents(hash.EventHashes{selfParent.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(2)
				e.SetLamport(1)
				e.SetID([24]byte{1})
				return ltypes.Events{e}
			}(),
			parentscheck.ErrWrongSeq, false},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(1)
			return e
		}(),
			nil,
			parentscheck.ErrWrongSeq, false},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(1)
			e.SetLamport(1)
			e.SetParents(hash.EventHashes{e.ID()})
			return e
		}(),
			nil,
			nil, true},
	}

	for _, tt := range tests {
		parentsCheck := parentscheck.New()
		if tt.wantPanic {
			assert.Panics(t, func() {
				err := parentsCheck.Validate(tt.e, tt.pe)
				if err != nil {
					return
				}
			})
		} else {
			assert.Equal(t, tt.wantErr, parentsCheck.Validate(tt.e, tt.pe))
		}
	}
}

func TestAllEventValidation(t *testing.T) {
	var tests = []struct {
		e       ltypes.Event
		pe      ltypes.Events
		wantErr error
	}{
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(2)
			e.SetParents(hash.EventHashes{e.ID()})
			return e
		}(),
			nil,
			basiccheck.ErrNotInited},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(1)
			e.SetLamport(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			return e
		}(),
			nil,
			epochcheck.ErrAuth},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(2)
			e.SetLamport(2)
			e.SetCreator(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			e.SetParents(hash.EventHashes{e.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(1)
				e.SetLamport(1)
				return ltypes.Events{e}
			}(),
			parentscheck.ErrWrongSelfParent},
		{func() ltypes.Event {
			e := &tdag.TestEvent{}
			e.SetSeq(1)
			e.SetLamport(2)
			e.SetCreator(1)
			e.SetEpoch(1)
			e.SetFrame(1)
			e.SetParents(hash.EventHashes{e.ID()})
			return e
		}(),
			func() ltypes.Events {
				e := &tdag.TestEvent{}
				e.SetSeq(1)
				e.SetLamport(1)
				return ltypes.Events{e}
			}(),
			nil},
	}

	tr := new(testReader)

	checkers := Checkers{
		Basiccheck:   basiccheck.New(),
		Epochcheck:   epochcheck.New(tr),
		Parentscheck: parentscheck.New(),
	}

	for _, tt := range tests {
		assert.Equal(t, tt.wantErr, checkers.Validate(tt.e, tt.pe))
	}
}
