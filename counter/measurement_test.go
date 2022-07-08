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
