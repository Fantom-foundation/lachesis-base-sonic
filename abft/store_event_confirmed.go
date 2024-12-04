package abft

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

// SetEventConfirmedOn stores confirmed event hash.
func (s *Store) SetEventConfirmedOn(e hash.EventHash, on idx.FrameID) {
	key := e.Bytes()

	if err := s.epochTable.ConfirmedEvent.Put(key, on.Bytes()); err != nil {
		s.crit(err)
	}
}

// GetEventConfirmedOn returns confirmed event hash.
func (s *Store) GetEventConfirmedOn(e hash.EventHash) idx.FrameID {
	key := e.Bytes()

	buf, err := s.epochTable.ConfirmedEvent.Get(key)
	if err != nil {
		s.crit(err)
	}
	if buf == nil {
		return 0
	}

	return idx.BytesToFrame(buf)
}
