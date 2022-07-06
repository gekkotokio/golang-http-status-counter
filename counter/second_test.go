package counter

import (
	"net/http"
	"testing"
)

func TestNewSecond(t *testing.T) {
	s := newSecond(http.StatusOK)

	if s.statuses[http.StatusOK].counter != 0 {
		t.Errorf("expected value was 0 but %v", s.statuses[http.StatusOK].counter)
	}

	s.statuses[http.StatusOK].increment()

	if s.statuses[http.StatusOK].counter != 1 {
		t.Errorf("expected value was 1 but %v", s.statuses[http.StatusOK].counter)
	}
}

func TestAddStatus(t *testing.T) {
	s := newSecond(http.StatusOK)

	if s.statuses[http.StatusOK].counter != 0 {
		t.Errorf("expected value was 0 but %v", s.statuses[http.StatusOK].counter)
	}

	if err := s.addStatus(http.StatusOK); err == nil {
		t.Error("expected returned error but nil")
	}

	if err := s.addStatus(http.StatusNotFound); err != nil {
		t.Errorf("expected returned no errors but %v", err.Error())
	}

	s.statuses[http.StatusOK].increment()

	if s.statuses[http.StatusOK].counter != 1 {
		t.Errorf("expected value was 1 but %v", s.statuses[http.StatusOK].counter)
	}

	if s.statuses[http.StatusNotFound].counter != 0 {
		t.Errorf("expected value was 0 but %v", s.statuses[http.StatusNotFound].counter)
	}

	s.statuses[http.StatusNotFound].increment()
	s.statuses[http.StatusNotFound].increment()

	if s.statuses[http.StatusNotFound].counter != 2 {
		t.Errorf("expected value was 2 but %v", s.statuses[http.StatusNotFound].counter)
	}

	if s.statuses[http.StatusOK].counter != 1 {
		t.Errorf("expected value was 1 but %v", s.statuses[http.StatusOK].counter)
	}
}

func TestSecondReset(t *testing.T) {
	second := newSecond(http.StatusContinue)

	testValidValues := []struct {
		statusCode int
		howMany    int
		expected   int
	}{
		{
			statusCode: http.StatusOK,
			howMany:    100,
			expected:   100,
		}, {
			statusCode: http.StatusNotFound,
			howMany:    150,
			expected:   150,
		}, {
			statusCode: http.StatusNotModified,
			howMany:    200,
			expected:   200,
		},
	}

	testInvalidValues := []struct {
		statusCode int
		howMany    int
		expected   int
	}{
		{
			statusCode: http.StatusBadRequest,
			howMany:    50,
			expected:   100,
		}, {
			statusCode: http.StatusUnauthorized,
			howMany:    150,
			expected:   100,
		}, {
			statusCode: http.StatusInternalServerError,
			howMany:    100,
			expected:   101,
		},
	}

	for _, test := range testValidValues {
		if err := second.addStatus(test.statusCode); err != nil {
			t.Errorf("expected returned no errors but %v", err.Error())
		}

		for i := 0; i < test.howMany; i++ {
			second.statuses[test.statusCode].increment()
		}

		if second.statuses[test.statusCode].counter != test.expected {
			t.Errorf("expected value was %v but %v", test.expected, second.statuses[http.StatusOK].counter)
		}
	}

	for _, test := range testInvalidValues {
		if err := second.addStatus(test.statusCode); err != nil {
			t.Errorf("expected returned no errors but %v", err.Error())
		}

		for i := 0; i < test.howMany; i++ {
			second.statuses[test.statusCode].increment()
		}

		if second.statuses[test.statusCode].counter == test.expected {
			t.Errorf("expected value was not equal to %v but equal to %v", test.expected, second.statuses[http.StatusOK].counter)
		}
	}

	second.resetWithLockContext()

	for statusCode, status := range second.statuses {
		if status.counter != 0 {
			t.Errorf("expected value was 0 of status code %v but %v", statusCode, status.counter)
		}
	}
}

func TestSecondResetWithLockContext(t *testing.T) {
	second := newSecond(http.StatusContinue)

	testValidValues := []struct {
		statusCode int
		howMany    int
		expected   int
	}{
		{
			statusCode: http.StatusOK,
			howMany:    100,
			expected:   100,
		}, {
			statusCode: http.StatusNotFound,
			howMany:    150,
			expected:   150,
		}, {
			statusCode: http.StatusNotModified,
			howMany:    200,
			expected:   200,
		},
	}

	testInvalidValues := []struct {
		statusCode int
		howMany    int
		expected   int
	}{
		{
			statusCode: http.StatusBadRequest,
			howMany:    50,
			expected:   100,
		}, {
			statusCode: http.StatusUnauthorized,
			howMany:    150,
			expected:   100,
		}, {
			statusCode: http.StatusInternalServerError,
			howMany:    100,
			expected:   101,
		},
	}

	for _, test := range testValidValues {
		if err := second.addStatus(test.statusCode); err != nil {
			t.Errorf("expected returned no errors but %v", err.Error())
		}

		for i := 0; i < test.howMany; i++ {
			second.statuses[test.statusCode].increment()
		}

		if second.statuses[test.statusCode].counter != test.expected {
			t.Errorf("expected value was %v but %v", test.expected, second.statuses[http.StatusOK].counter)
		}
	}

	for _, test := range testInvalidValues {
		if err := second.addStatus(test.statusCode); err != nil {
			t.Errorf("expected returned no errors but %v", err.Error())
		}

		for i := 0; i < test.howMany; i++ {
			second.statuses[test.statusCode].increment()
		}

		if second.statuses[test.statusCode].counter == test.expected {
			t.Errorf("expected value was not equal to %v but equal to %v", test.expected, second.statuses[http.StatusOK].counter)
		}
	}

	second.resetWithLockContext()

	for statusCode, status := range second.statuses {
		if status.counter != 0 {
			t.Errorf("expected value was 0 of status code %v but %v", statusCode, status.counter)
		}
	}
}

func BenchmarkResetStatusCounterLockedEachStruct(b *testing.B) {
	second := newSecond(http.StatusContinue)

	testValues := []struct {
		statusCode int
		howMany    int
	}{
		{
			statusCode: http.StatusOK,
			howMany:    1000,
		}, {
			statusCode: http.StatusNotFound,
			howMany:    1000,
		}, {
			statusCode: http.StatusNotModified,
			howMany:    1000,
		}, {
			statusCode: http.StatusBadRequest,
			howMany:    1000,
		}, {
			statusCode: http.StatusUnauthorized,
			howMany:    1000,
		}, {
			statusCode: http.StatusInternalServerError,
			howMany:    1000,
		},
	}

	for i := 0; i < b.N; i++ {
		for _, test := range testValues {
			second.addStatus(test.statusCode)
		}

		second.reset()
	}
}

func BenchmarkResetStatusCounterLockedWholeStruct(b *testing.B) {
	second := newSecond(http.StatusContinue)

	testValues := []struct {
		statusCode int
		howMany    int
	}{
		{
			statusCode: http.StatusOK,
			howMany:    1000,
		}, {
			statusCode: http.StatusNotFound,
			howMany:    1000,
		}, {
			statusCode: http.StatusNotModified,
			howMany:    1000,
		}, {
			statusCode: http.StatusBadRequest,
			howMany:    1000,
		}, {
			statusCode: http.StatusUnauthorized,
			howMany:    1000,
		}, {
			statusCode: http.StatusInternalServerError,
			howMany:    1000,
		},
	}

	for i := 0; i < b.N; i++ {
		for _, test := range testValues {
			second.addStatus(test.statusCode)
		}

		second.resetWithLockContext()
	}
}
