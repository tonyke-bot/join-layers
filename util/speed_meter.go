package util

import (
	"sync"
	"time"
)

type speedMetric struct {
	Prev      *speedMetric
	Timestamp time.Time
	Data      uint32
}

type ProgressMeter struct {
	lastMetric *speedMetric
	counter    uint32
	max        uint32
	window     time.Duration
	lock       sync.Mutex
}

func NewProgressMeter(window time.Duration, max uint32) *ProgressMeter {
	return &ProgressMeter{window: window, max: max}
}

func (s *ProgressMeter) purge() {
	expiration := time.Now().Add(-s.window)

	if s.lastMetric == nil {
		return
	}

	if s.lastMetric.Timestamp.Before(expiration) {
		s.lastMetric = nil
		return
	}

	cursor := s.lastMetric
	for ; cursor != nil && cursor.Timestamp.After(expiration); cursor = cursor.Prev {
	}

	for cursor != nil {
		prev := cursor.Prev
		cursor.Prev = nil
		cursor = prev
	}
}

func (s *ProgressMeter) Log() *ProgressMeter {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.purge()

	s.counter++
	s.lastMetric = &speedMetric{
		Prev:      s.lastMetric,
		Timestamp: time.Now(),
		Data:      s.counter,
	}

	return s
}

func (s *ProgressMeter) Finished() bool {
	return s.Current() >= s.max
}

func (s *ProgressMeter) Current() uint32 {
	lastMetric := s.lastMetric
	if lastMetric == nil {
		return 0
	}

	return lastMetric.Data
}

func (s *ProgressMeter) ETA() time.Duration {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.purge()

	if s.lastMetric == nil || s.lastMetric.Prev == nil {
		return time.Duration(0)
	}

	max := s.lastMetric.Data
	min := max
	endTime := s.lastMetric.Timestamp
	startTime := endTime

	cursor := s.lastMetric
	for cursor != nil {
		min = cursor.Data
		startTime = cursor.Timestamp
		cursor = cursor.Prev
	}

	duration := endTime.Sub(startTime)
	speed := float64(max-min) / float64(duration.Milliseconds())

	left := s.max - max
	return time.Duration(float64(left) / speed * float64(time.Millisecond))
}
