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

		if len(s.status) != 1 {
			t.Errorf("expected length of initialized struct was 1 but %v", len(s.status))
		} else if s.status[status.code] != status.expected {
			t.Errorf("expected %v were counted %v but %v", status.code, status.expected, s.getCounterWithLockContext(status.code))
		} else if len(s.status) != 1 {
			t.Errorf("expected lenght was 1 but %v", len(s.status))
		} else if s.getCounterWithLockContext(status.code) != 0 {
			t.Errorf("expected 0 but returned %v", s.getCounterWithLockContext(status.code))
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
			t.Errorf("expected %v were counted %v but %v", status.code, status.expected, s.getCounterWithLockContext(status.code))
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

	p := newStatuses(http.StatusInternalServerError)

	if len(p.status) != 2 {
		t.Errorf("expected length of the buffered struct was 2 but %v", len(p.status))
	}
}

func BenchmarkInitilaizeEachtime(b *testing.B) {
	max := 30
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := make(map[int]int, 1)
		m[-1] = 0
		s := statuses{status: m}

		for j := 0; j < max; j++ {
			s.status[j] = 1
		}
	}
}

func BenchmarkPooled(b *testing.B) {
	max := 30
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := newStatuses(-1)

		for j := 0; j < max; j++ {
			s.status[j] = 1
		}

		s.resetWithLockContext()
	}
}
