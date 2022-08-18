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

// CountUpWithLockContext increases the number of the given HTTP status codes with thread-safe.
func (m *Measurement) CountUpWithLockContext(statusCode int) {
	epoch := time.Now().Unix()

	m.withLockContext(func() {
		if _, ok := m.at[epoch]; !ok {
			m.addNewEpochRecord(epoch, statusCode)
		}
	})

	m.at[epoch].incrementWithLockContext(statusCode)
}

// LatestRecordedAt returns latest recorded timestamp.
func (m *Measurement) LatestRecordedAt() (epoch int64) {
	epoch = 0

	m.withLockContext(func() {
		for timestamp, _ := range m.at {
			if epoch < timestamp {
				epoch = timestamp
			}
		}
	})

	return epoch
}

// OldestRecordedAt returns oldest recorded timestamp with thread-safe.
func (m *Measurement) OldestRecordedAt() (epoch int64) {
	epoch = time.Now().Unix()

	m.withLockContext(func() {
		for timestamp, _ := range m.at {
			if timestamp < epoch {
				epoch = timestamp
			}
		}
	})

	return epoch
}

// Length returns the seconds how log HTTP status codes are recorded with thread-safe.
func (m *Measurement) LengthWithLockContext() (duration int) {
	duration = 0

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

// ExpireRecordsWithLockContext deletes the records older than the given UNIX time with thread-safe.
// otherwise returned error if there were no expired records.
func (m *Measurement) ExpireRecordsWithLockContext(expiredBefore int64) error {
	var e error

	m.withLockContext(func() {
		e = m.expireRecords(expiredBefore)
	})

	return e
}

func (m *Measurement) getRecordsAt(epoch int64) (map[int]int, error) {
	records := make(map[int]int)

	if _, ok := m.at[epoch]; !ok {
		return records, fmt.Errorf("records at %v not found", epoch)
	} else {
		for code, counter := range m.at[epoch].status {
			records[code] = counter
		}
	}

	return records, nil
}

// GetRecordsAt returns the counters of status codes for the given UNIX time with thread-safe.
func (m *Measurement) GetRecordsAtWithLockContext(epoch int64) (records map[int]int, err error) {
	records = make(map[int]int)

	m.withLockContext(func() {
		records, err = m.getRecordsAt(epoch)
	})

	return records, err
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

// Extract returns the numbers of HTTP status codes by seconds between the given ranges  with thread-safe.
// Target range is fromEpoch <= range < toEpoch.
func (m *Measurement) ExtractWithLockContext(fromEpoch int64, toEpoch int64) (record Record, err error) {
	record = make(map[int64]map[int]int)

	m.withLockContext(func() {
		record, err = m.extract(fromEpoch, toEpoch)
	})

	return record, err
}

func (m *Measurement) addNewEpochRecord(epoch int64, statusCode int) {
	m.at[epoch] = newStatuses(statusCode)
}

func (m *Measurement) sumByStatusCodes() map[int]int {
	r := make(map[int]int)

	for _, statuses := range m.at {
		for code, counter := range statuses.status {
			if _, ok := r[code]; !ok {
				r[code] = 0
			}

			r[code] += counter
		}
	}

	return r
}

// SumByStatusCodesWithLockContext returns a map that sums the number of status codes per second with thread-safe.
func (m *Measurement) SumByStatusCodesWithLockContext() map[int]int {
	record := make(map[int]int)

	m.withLockContext(func() {
		record = m.sumByStatusCodes()
	})

	return record
}

func (m *Measurement) withLockContext(fn func()) {
	m.Lock()
	defer m.Unlock()

	fn()
}
