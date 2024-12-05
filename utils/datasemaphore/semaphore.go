package datasemaphore

import (
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

type DataSemaphore struct {
	processing    ltypes.Metric
	maxProcessing ltypes.Metric

	mu   sync.Mutex
	cond *sync.Cond

	warning func(received ltypes.Metric, processing ltypes.Metric, releasing ltypes.Metric)
}

func New(maxProcessing ltypes.Metric, warning func(received ltypes.Metric, processing ltypes.Metric, releasing ltypes.Metric)) *DataSemaphore {
	s := &DataSemaphore{
		maxProcessing: maxProcessing,
		warning:       warning,
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *DataSemaphore) Acquire(weight ltypes.Metric, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	s.mu.Lock()
	defer s.mu.Unlock()
	for !s.tryAcquire(weight) {
		if weight.Size > s.maxProcessing.Size || weight.Num > s.maxProcessing.Num || time.Now().After(deadline) {
			return false
		}
		s.cond.Wait()
	}
	return true
}

func (s *DataSemaphore) TryAcquire(weight ltypes.Metric) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tryAcquire(weight)
}

func (s *DataSemaphore) tryAcquire(metric ltypes.Metric) bool {
	tmp := s.processing
	tmp.Num += metric.Num
	tmp.Size += metric.Size
	if tmp.Num > s.maxProcessing.Num || tmp.Size > s.maxProcessing.Size {
		return false
	}
	s.processing = tmp
	return true
}

func (s *DataSemaphore) Release(weight ltypes.Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.processing.Num < weight.Num || s.processing.Size < weight.Size {
		if s.warning != nil {
			s.warning(s.processing, s.processing, weight)
		}
		s.processing = ltypes.Metric{}
	} else {
		s.processing.Num -= weight.Num
		s.processing.Size -= weight.Size
	}
	s.cond.Broadcast()
}

func (s *DataSemaphore) Terminate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxProcessing = ltypes.Metric{}
	s.cond.Broadcast()
}

func (s *DataSemaphore) Processing() ltypes.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.processing
}

func (s *DataSemaphore) Available() ltypes.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ltypes.Metric{
		Num:  s.maxProcessing.Num - s.processing.Num,
		Size: s.maxProcessing.Size - s.processing.Size,
	}
}
