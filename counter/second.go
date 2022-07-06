package counter

import (
	"fmt"
	"sync"
)

type second struct {
	statuses map[int]*status
}

var secondMutex sync.Mutex

func newSecond(statusCode int) second {
	s := newStatus()
	m := map[int]*status{statusCode: &s}

	return second{statuses: m}
}

func (s *second) addStatus(statusCode int) error {
	if _, ok := s.statuses[statusCode]; ok {
		return fmt.Errorf("already had status code key: %v", statusCode)
	}

	status := newStatus()
	s.statuses[statusCode] = &status

	return nil
}

func (s *second) reset() {
	for _, status := range s.statuses {
		status.resetWithLockContext()
	}
}

func secondWithLockContext(fn func()) {
	secondMutex.Lock()
	defer secondMutex.Unlock()

	fn()
}

// resetWithLockContext() runs faster than reset()
func (s *second) resetWithLockContext() {
	secondWithLockContext(func() {
		for _, status := range s.statuses {
			status.reset()
		}
	})
}
