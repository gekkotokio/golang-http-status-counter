package counter

import "sync"

type status struct {
	sync.Mutex
	counter int
}

func newStatus() status {
	return status{
		counter: 0,
	}
}

func (s *status) increment() {
	s.Lock()
	defer s.Unlock()
	s.counter++
}

func (s *status) reset() {
	s.Lock()
	defer s.Unlock()
	s.counter = 0
}
