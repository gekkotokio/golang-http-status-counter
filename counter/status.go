package counter

import "sync"

type status struct {
	mu      sync.Mutex
	counter int
}

func newStatus() status {
	return status{
		counter: 0,
	}
}

func (s *status) increment() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
}

func (s *status) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter = 0
}
