package counter

import "sync"

type status struct {
	sync.Mutex
	counter int
}

func newStatus() *status {
	return &status{
		counter: 0,
	}
}

func (s *status) increment() {
	s.counter++
}

func (s *status) reset() {
	s.counter = 0
}

func (s *status) withLockContext(fn func()) {
	s.Lock()
	defer s.Unlock()

	fn()
}

func (s *status) incrementWithLockContext() {
	s.withLockContext(func() {
		s.increment()
	})
}

func (s *status) resetWithLockContext() {
	s.withLockContext(func() {
		s.reset()
	})
}
