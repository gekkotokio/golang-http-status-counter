package counter

import (
	"fmt"
	"sync"
)

type second struct {
	sync.Mutex
	statuses map[int]*status
}

func newSecond(statusCode int) *second {
	s := newStatus()
	m := map[int]*status{statusCode: s}

	return &second{statuses: m}
}

func (s *second) addStatus(statusCode int) error {
	if s.hasStatusCode(statusCode) {
		return fmt.Errorf("already had status code key: %v", statusCode)
	}

	s.statuses[statusCode] = newStatus()

	return nil
}

func (s *second) countUp(statusCode int) {
	if !s.hasStatusCode(statusCode) {
		s.insertStatusCodeWithLockContext(statusCode)
	}

	s.statuses[statusCode].incrementWithLockContext()
}

func (s *second) hasStatusCode(statusCode int) bool {
	_, ok := s.statuses[statusCode]
	return ok
}

func (s *second) insertStatusCode(statusCode int) {
	s.statuses[statusCode] = newStatus()
}

func (s *second) insertStatusCodeWithLockContext(statusCode int) {
	s.withLockContext(func() {
		if !s.hasStatusCode(statusCode) {
			s.insertStatusCode(statusCode)
		}
	})
}

func (s *second) reset() {
	for _, status := range s.statuses {
		status.resetWithLockContext()
	}
}

func (s *second) withLockContext(fn func()) {
	s.Lock()
	defer s.Unlock()

	fn()
}

// resetWithLockContext() runs faster than reset()
func (s *second) resetWithLockContext() {
	s.withLockContext(func() {
		for _, status := range s.statuses {
			status.reset()
		}
	})
}
