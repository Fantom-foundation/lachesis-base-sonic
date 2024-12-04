package abft

import (
	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// SetEventConfirmedOn stores confirmed event ltypes.
func (s *Store) SetEventConfirmedOn(e ltypes.EventHash, on ltypes.FrameID) {
	key := e.Bytes()

	if err := s.epochTable.ConfirmedEvent.Put(key, on.Bytes()); err != nil {
		s.crit(err)
	}
}

// GetEventConfirmedOn returns confirmed event ltypes.
func (s *Store) GetEventConfirmedOn(e ltypes.EventHash) ltypes.FrameID {
	key := e.Bytes()

	buf, err := s.epochTable.ConfirmedEvent.Get(key)
	if err != nil {
		s.crit(err)
	}
	if buf == nil {
		return 0
	}

	return ltypes.BytesToFrameID(buf)
}
