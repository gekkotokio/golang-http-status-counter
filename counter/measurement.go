package counter

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Measurement struct {
	sync.Mutex
	at map[int64]*statuses
}

type Record map[int64]map[int]int

// NewMeasurement returns initialized Measurement struct.
// It has the counter of HTTP 200 status code of the UNIX time when initialized.
func NewMeasurement() Measurement {
	s := newStatuses(http.StatusOK)
	t := time.Now().Unix()

	return Measurement{
		at: map[int64]*statuses{t: s},
	}
}

// CountUp increases the number of the given HTTP status codes with thread-safe.
func (m *Measurement) CountUp(statusCode int) {
	epoch := time.Now().Unix()

	if !m.isRecordedAt(epoch) {
		m.insertSecondWithLockContext(epoch, statusCode)
	}

	m.at[epoch].incrementWithLockContext(statusCode)
}

// LatestRecordedAt returns latest recorded timestamp.
func (m *Measurement) LatestRecordedAt() int64 {
	var t int64 = 0

	m.withLockContext(func() {
		for timestamp, _ := range m.at {
			if t < timestamp {
				t = timestamp
			}
		}
	})

	return t
}

// OldestRecordedAt returns oldest recorded timestamp.
func (m *Measurement) OldestRecordedAt() int64 {
	t := time.Now().Unix()

	m.withLockContext(func() {
		for timestamp, _ := range m.at {
			if timestamp < t {
				t = timestamp
			}
		}
	})

	return t
}

// RecordedDuration returns the seconds how log HTTP status codes are recorded.
func (m *Measurement) RecordedDuration() int {
	duration := 0

	m.withLockContext(func() {
		duration = len(m.at)
	})

	return duration
}

func (m *Measurement) expireRecords(expiredBefore int64) error {
	removed := false

	for epoch, _ := range m.at {
		if epoch < expiredBefore {
			m.at[epoch].resetWithLockContext()
			delete(m.at, epoch)

			if !removed {
				removed = true
			}
		}
	}

	if !removed {
		return fmt.Errorf("there were no expired records older than %v", expiredBefore)
	}

	return nil
}

// ExpireRecordsWithLockContext deletes the records older than the given UNIX time.
// otherwise returned error if there were no expired records.
func (m *Measurement) ExpireRecordsWithLockContext(expiredBefore int64) error {
	var e error

	m.withLockContext(func() {
		e = m.expireRecords(expiredBefore)
	})

	return e
}

// GetRecordsAt returns the counters of status codes for the given UNIX time.
func (m *Measurement) GetRecordsAt(at int64) (map[int]int, error) {
	if _, ok := m.at[at]; !ok {
		return nil, fmt.Errorf("records at %v not found", at)
	}

	return m.at[at].status, nil
}

func (m *Measurement) extract(fromEpoch int64, toEpoch int64) (Record, error) {
	if fromEpoch < 1 {
		return nil, fmt.Errorf("fromEpoch should be more than 1")
	} else if toEpoch < 1 {
		return nil, fmt.Errorf("toEpoch should be more than 1")
	} else if toEpoch <= fromEpoch {
		return nil, fmt.Errorf("toEpoch should be greater than fromEpoch")
	}

	r := make(map[int64]map[int]int)

	for epoch, responses := range m.at {
		if fromEpoch <= epoch && epoch < toEpoch {
			if _, ok := r[epoch]; !ok {
				r[epoch] = make(map[int]int)
			}

			for statusCode, counter := range responses.status {
				r[epoch][statusCode] = counter
			}
		}
	}

	if len(r) == 0 {
		return nil, fmt.Errorf("the given values were out of range: %v to %v", fromEpoch, toEpoch)
	}

	return r, nil
}

// Extract returns the numbers of HTTP status codes by seconds between the given ranges in thread-safe.
// Target range is fromEpoch <= range < toEpoch.
func (m *Measurement) ExtractWithLockContext(fromEpoch int64, toEpoch int64) (Record, error) {
	var e error
	r := make(map[int64]map[int]int)

	m.withLockContext(func() {
		r, e = m.extract(fromEpoch, toEpoch)
	})

	return r, e
}

func (m *Measurement) addStatusCodeRecord(epoch int64, statusCode int) {
	m.at[epoch] = newStatuses(statusCode)
}

func (m *Measurement) isRecordedAt(epoch int64) bool {
	_, ok := m.at[epoch]
	return ok
}

func (m *Measurement) insertSecondWithLockContext(epoch int64, statusCode int) {
	m.withLockContext(func() {
		// double-check the given key is empty
		if !m.isRecordedAt(epoch) {
			m.addStatusCodeRecord(epoch, statusCode)
		}
	})
}

func (m *Measurement) withLockContext(fn func()) {
	m.Lock()
	defer m.Unlock()

	fn()
}
