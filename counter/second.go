package counter

import "fmt"

type second struct {
	statuses map[int]*status
}

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
		status.reset()
	}
}
