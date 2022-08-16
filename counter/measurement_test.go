package counter

import (
	"net/http"
	"testing"
	"time"
)

func TestNewMeasurement(t *testing.T) {
	m := NewMeasurement()

	if len(m.at) != 1 {
		t.Errorf("expected length of Measurement.period was 1 but %v", len(m.at))
	}
}

// TestCountUp would be failed because it uses elapsed time for the given values.
// Depending on the execution time of tests, it may not be possible to get the expected results.
func TestCountUp(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()
	later := now + 1

	if m.at[now].getCounterWithLockContext(http.StatusOK) != 0 {
		t.Errorf("expected value was 0 but %v", m.at[now].getCounterWithLockContext(http.StatusOK))
	}

	m.CountUp(http.StatusOK)

	if m.at[now].getCounterWithLockContext(http.StatusOK) != 1 {
		t.Errorf("expected value was 1 but %v", m.at[now].getCounterWithLockContext(http.StatusOK))
	}

	if _, ok := m.at[now].status[http.StatusNotModified]; ok {
		t.Error("expected returned false but true")
	}

	m.CountUp(http.StatusNotModified)

	if m.at[now].getCounterWithLockContext(http.StatusNotModified) != 1 {
		t.Errorf("expected value was 1 but %v", m.at[now].getCounterWithLockContext(http.StatusNotModified))
	}

	time.Sleep(time.Second + 1)

	m.CountUp(http.StatusNotModified)

	if m.at[now].getCounterWithLockContext(http.StatusNotModified) != 1 {
		t.Errorf("expected value was 1 but %v", m.at[now].getCounterWithLockContext(http.StatusNotModified))
	}

	if _, ok := m.at[later]; !ok {
		t.Errorf("expected returned true but %v", ok)
	}

	m.CountUp(http.StatusNotModified)

	if m.at[later].getCounterWithLockContext(http.StatusNotModified) != 2 {
		t.Errorf("expected value was 2 but %v", m.at[later].getCounterWithLockContext(http.StatusNotModified))
	}
}

func TestIsRecordedAt(t *testing.T) {
	m := NewMeasurement()
	now := time.Now().Unix()
	testEpoch := now + 1

	if m.isRecordedAt(testEpoch) {
		t.Errorf("expected returned false but true")
		for epoch, _ := range m.at {
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

	if m.at[now].getCounterWithLockContext(http.StatusNotFound) != 0 {
		t.Errorf("expected value was 0 but %v", m.at[now].getCounterWithLockContext(http.StatusNotFound))
	}
}

func TestExtract(t *testing.T) {
	duration := 305
	m := NewMeasurement()
	max := time.Now().Unix()
	min := max - int64(duration)
	ranged := max - 300

	for i := 0; i < duration; i++ {
		m.insertSecondWithLockContext(min, http.StatusNotModified)
		m.addStatusCodeRecord(min, http.StatusNotFound)
		m.at[min].status[http.StatusNotModified] = 100
		m.at[min].status[http.StatusNotFound] = 10
		min++
	}

	// doing NewMeasurement() generates one length of Measurement struct.
	// and adds the size of 305 seconds of structs.
	if m.RecordedDuration() != (duration + 1) {
		t.Errorf("expected the length of m was %v but %v", duration, len(m.at))
	}

	if m.OldestRecordedAt() != (max - int64(duration)) {
		t.Errorf("expected oldest record was recorded at %v but %v", (max - int64(duration)), m.OldestRecordedAt())
	}

	if m.LatestRecordedAt() != max {
		t.Errorf("expected latest record was recorded at %v but %v", max, m.LatestRecordedAt())
	}

	for epoch, statuses := range m.at {
		if records, err := m.GetRecordsAt(epoch); err != nil {
			t.Errorf("expected no errors but returned error: %v", err.Error())
		} else {
			for code, counter := range statuses.status {
				if records[code] != counter {
					t.Errorf("expected %v counter was %v but %v", code, counter, records[code])
				}
			}
		}
	}

	// Test of thresholds to extranct records
	var from int64 = 0
	var to int64 = 0

	if _, err := m.ExtractWithLockContext(from, to); err == nil {
		t.Error("expected returned error but no error")
	} else if _, err := m.ExtractWithLockContext(from+1, to); err == nil {
		t.Error("expected returned error but no error")
	} else if _, err := m.ExtractWithLockContext(from+1, to+1); err == nil {
		t.Error("expected returned error but no error")
	} else if _, err := m.ExtractWithLockContext(from+2, to+1); err == nil {
		t.Error("expected returned error but no error")
	} else if _, err := m.ExtractWithLockContext(from+1, to+2); err == nil {
		t.Error("expected returned error but no error")
	} else if _, err := m.GetRecordsAt(from); err == nil {
		t.Error("expected returned error but no error")
	}

	if r, err := m.ExtractWithLockContext(ranged, max); err != nil {
		t.Errorf("expected no errors but %v", err.Error())
	} else if len(r) != 300 {
		t.Errorf("expected length was 300 but %v", len(r))
	} else if _, ok := r[max]; ok {
		t.Errorf("expected returned false but %v", ok)
	} else {
		counter := 0

		for _, codes := range r {
			counter += codes[http.StatusNotModified]
		}

		if counter != (100 * 300) {
			t.Errorf("expected values was 30000 but %v", counter)
		}
	}

	min = max - int64(duration)

	if err := m.ExpireRecordsWithLockContext(from); err == nil {
		t.Error("expected returned errors but no errors.")
	} else if err := m.ExpireRecordsWithLockContext(min); err == nil {
		t.Error("expected returned errors but no errors.")
	} else if err := m.ExpireRecordsWithLockContext(min + 1); err != nil {
		t.Errorf("expected no errors but returned errors: %v", err.Error())
	}
}
