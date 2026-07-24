package scan

import "sync"

type weightedScheduler struct {
	mu        sync.Mutex
	condition *sync.Cond
	available int
	capacity  int
}

func newWeightedScheduler(capacity int) *weightedScheduler {
	if capacity < 1 {
		capacity = 1
	}
	scheduler := &weightedScheduler{available: capacity, capacity: capacity}
	scheduler.condition = sync.NewCond(&scheduler.mu)
	return scheduler
}

func (s *weightedScheduler) acquire(weight int) {
	weight = s.normalizedWeight(weight)
	s.mu.Lock()
	defer s.mu.Unlock()
	for s.available < weight {
		s.condition.Wait()
	}
	s.available -= weight
}

func (s *weightedScheduler) release(weight int) {
	weight = s.normalizedWeight(weight)
	s.mu.Lock()
	s.available += weight
	if s.available > s.capacity {
		s.available = s.capacity
	}
	s.mu.Unlock()
	s.condition.Broadcast()
}

func (s *weightedScheduler) normalizedWeight(weight int) int {
	if weight < 1 {
		return 1
	}
	if weight > s.capacity {
		return s.capacity
	}
	return weight
}
