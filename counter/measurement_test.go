package counter

import (
	"net/http"
	"testing"
	"time"
)

func TestNewMeasurement(t *testing.T) {
	m := NewMeasurement()

	if len(m.period) != 1 {
		t.Errorf("expected length of Measurement.period was 1 but %v", len(m.period))
	}
}

// TestCountUp would be failed because it uses elapsed time for the given values.
// Depending on the execution time of tests, it may not be possible to get the expected results.
func TestCountUp(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()
	later := now + 1

	if m.period[now].statuses[http.StatusOK].counter != 0 {
		t.Errorf("expected value was 0 but %v", m.period[now].statuses[http.StatusOK].counter)
	}

	m.CountUp(http.StatusOK)

	if m.period[now].statuses[http.StatusOK].counter != 1 {
		t.Errorf("expected value was 1 but %v", m.period[now].statuses[http.StatusOK].counter)
	}

	if _, ok := m.period[now].statuses[http.StatusNotModified]; ok {
		t.Error("expected returned false but true")
	}

	m.CountUp(http.StatusNotModified)

	if m.period[now].statuses[http.StatusNotModified].counter != 1 {
		t.Errorf("expected value was 1 but %v", m.period[now].statuses[http.StatusNotModified].counter)
	}

	time.Sleep(time.Second + 1)

	m.CountUp(http.StatusNotModified)

	if m.period[later].statuses[http.StatusNotModified].counter != 1 {
		t.Errorf("expected value was 1 but %v", m.period[now].statuses[http.StatusNotModified].counter)
	}

	if _, ok := m.period[later]; !ok {
		t.Errorf("expected returned true but %v", ok)
	}

	m.CountUp(http.StatusNotModified)

	if m.period[later].statuses[http.StatusNotModified].counter != 2 {
		t.Errorf("expected value was 2 but %v", m.period[now].statuses[http.StatusNotModified].counter)
	}
}

func TestAddStatusCodeRecord(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()

	m.addStatusCodeRecord(now, http.StatusServiceUnavailable)

	if checked := m.hasStatusCodeRecord(now, http.StatusServiceUnavailable); !checked {
		t.Errorf("expected returned true but %v", checked)
	}
}

func TestHasStatusCodeRecord(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()

	if checked := m.hasStatusCodeRecord(now, http.StatusAccepted); checked {
		t.Errorf("expected returned false but %v", checked)
	}
}

func TestIsRecordedAt(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()
	testEpoch := now + 1

	if m.isRecordedAt(testEpoch) {
		t.Errorf("expected returned false but true")
		for epoch, _ := range m.period {
			t.Errorf("initialized epoch was %v, test value was %v", epoch, testEpoch)
		}
	} else if !m.isRecordedAt(now) {
		t.Errorf("initialized epoch was %v, test value was %v", now, testEpoch)
	}
}

func TestInsertSecondWithLockContext(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()
	new := now + 1

	m.insertSecondWithLockContext(new, http.StatusNotFound)

	/*for epoch, statuses := range m.period {
		for statusCode, counter := range statuses.statuses {
			t.Errorf("%v: %v: %v", epoch, statusCode, counter.counter)
		}
	}*/

	if m.period[new].statuses[http.StatusNotFound].counter != 0 {
		t.Errorf("expected value was 0 but %v", m.period[new].statuses[http.StatusNotFound].counter)
	}
}

func TestInsertStatusCodeRecordWithLockContext(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()

	m.insertStatusCodeRecordWithLockContext(now, http.StatusNotModified)

	if _, ok := m.period[now].statuses[http.StatusNotModified]; !ok {
		t.Errorf("expected returned false but %v", ok)
	}
}
