package counter

import "sync"

type statuses struct {
	sync.Mutex
	status map[int]int
}

func newStatuses(statusCode int) *statuses {
	s := map[int]int{statusCode: 0}

	return &statuses{
		status: s,
	}
}

func (s *statuses) getCounter(statusCode int) int {
	if _, ok := s.status[statusCode]; !ok {
		return 0
	}

	return s.status[statusCode]
}

func (s *statuses) getCounterWithLockContext(statusCode int) int {
	c := 0

	s.withLockContext(func() {
		c = s.getCounter(statusCode)
	})

	return c
}

func (s *statuses) increment(statusCode int) {
	if _, ok := s.status[statusCode]; !ok {
		s.status[statusCode] = 1
	} else {
		s.status[statusCode]++
	}
}

func (s *statuses) incrementWithLockContext(statusCode int) {
	s.withLockContext(func() {
		s.increment(statusCode)
	})
}

func (s *statuses) reset() {
	for code := range s.status {
		s.status[code] = 0
	}
}

func (s *statuses) resetWithLockContext() {
	s.withLockContext(func() {
		s.reset()
	})
}

func (s *statuses) withLockContext(fn func()) {
	s.Lock()
	defer s.Unlock()

	fn()
}
