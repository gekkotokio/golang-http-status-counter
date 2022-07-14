package counter

import (
	"net/http"
	"testing"
)

func TestNewStatuses(t *testing.T) {
	success := []struct {
		code     int
		expected int
	}{
		{
			code:     http.StatusOK,
			expected: 0,
		}, {
			code:     http.StatusNotModified,
			expected: 0,
		}, {
			code:     http.StatusNotFound,
			expected: 0,
		},
	}

	for _, status := range success {
		s := newStatuses(status.code)

		if s.status[status.code] != status.expected {
			t.Errorf("expected %v were counted %v but %v", status.code, status.expected, s.status[status.code])
		} else if len(s.status) != 1 {
			t.Errorf("expected lenght was 1 but %v", len(s.status))
		}
	}

	failed := []struct {
		code     int
		expected int
	}{
		{
			code:     http.StatusOK,
			expected: 1,
		}, {
			code:     http.StatusNotModified,
			expected: 1,
		}, {
			code:     http.StatusNotFound,
			expected: 1,
		},
	}

	for _, status := range failed {
		s := newStatuses(status.code)

		if s.status[status.code] == status.expected {
			t.Errorf("expected %v were counted %v but %v", status.code, status.expected, s.status[status.code])
		} else if len(s.status) != 1 {
			t.Errorf("expected lenght was 1 but %v", len(s.status))
		}
	}
}

func TestIncrementCounterWithLockContext(t *testing.T) {
	s := newStatuses(http.StatusInternalServerError)

	if c := s.getCounterWithLockContext(http.StatusInternalServerError); c != 0 {
		t.Errorf("expected returned 0 but %v", c)
	}

	s.incrementWithLockContext(http.StatusInternalServerError)

	if c := s.getCounterWithLockContext(http.StatusInternalServerError); c != 1 {
		t.Errorf("expected returned 1 but %v", c)
	}

	if c := s.getCounterWithLockContext(http.StatusNotFound); c != 0 {
		t.Errorf("expected returned 0 but %v", c)
	}

	s.incrementWithLockContext(http.StatusNotFound)

	if c := s.getCounterWithLockContext(http.StatusNotFound); c != 1 {
		t.Errorf("expected returned 1 but %v", c)
	}
}

func TestResetCounterWithLockContext(t *testing.T) {
	s := newStatuses(http.StatusInternalServerError)
	max := 100

	for i := 0; i < max; i++ {
		s.incrementWithLockContext(http.StatusInternalServerError)
	}

	if c := s.getCounterWithLockContext(http.StatusInternalServerError); c != max {
		t.Errorf("expected returned %v but %v", max, c)
	}

	if c := s.getCounterWithLockContext(http.StatusNotFound); c != 0 {
		t.Errorf("expected returned 0 but %v", c)
	}

	s.incrementWithLockContext(http.StatusNotFound)

	s.resetWithLockContext()

	if c := s.getCounterWithLockContext(http.StatusInternalServerError); c != 0 {
		t.Errorf("expected returned 0 but %v", c)
	} else if c := s.getCounterWithLockContext(http.StatusNotFound); c != 0 {
		t.Errorf("expected returned 0 but %v", c)
	}
}
