package counter

import "sync"

var statusMutex sync.Mutex

type status struct {
	counter int
}

func newStatus() status {
	return status{
		counter: 0,
	}
}

func (s *status) increment() {
	s.counter++
}

func (s *status) reset() {
	s.counter = 0
}

func statusWithLockContext(fn func()) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	fn()
}

func (s *status) incrementWithLockContext() {
	statusWithLockContext(func() {
		s.increment()
	})
}

func (s *status) resetWithLockContext() {
	statusWithLockContext(func() {
		s.reset()
	})
}
