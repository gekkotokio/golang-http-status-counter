package counter

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Measurement struct {
	sync.Mutex
	period map[int64]*second
}

type Record map[int64]map[int]int

// NewMeasurement returns initialized Measurement struct.
// It has the counter of HTTP 200 status code of the epoch time when initialized.
func NewMeasurement() Measurement {
	s := newSecond(http.StatusOK)
	t := time.Now().Unix()
	m := map[int64]*second{t: s}

	return Measurement{period: m}
}

// CountUp increases the number of the given HTTP status codes with thread-safe.
func (m *Measurement) CountUp(statusCode int) {
	epoch := time.Now().Unix()

	if !m.isRecordedAt(epoch) {
		m.insertSecondWithLockContext(epoch, statusCode)
	}

	if !m.hasStatusCodeRecord(epoch, statusCode) {
		m.insertStatusCodeRecordWithLockContext(epoch, statusCode)
	}

	m.period[epoch].statuses[statusCode].incrementWithLockContext()
}

// Extract returns the numbers of HTTP status codes by seconds between the given ranges.
// Range target is fromEpoch <= range < toEpoch.
func (m *Measurement) Extract(fromEpoch int64, toEpoch int64) (Record, error) {
	if fromEpoch < 1 {
		return nil, fmt.Errorf("fromEpoch should be more than 1")
	} else if toEpoch < 1 {
		return nil, fmt.Errorf("toEpoch should be more than 1")
	} else if toEpoch <= fromEpoch {
		return nil, fmt.Errorf("toEpoch should be greater than toEpoch")
	}

	r := make(map[int64]map[int]int)

	for epoch, responses := range m.period {
		if epoch < toEpoch && fromEpoch <= epoch {
			r[epoch] = make(map[int]int)

			for code, status := range responses.statuses {
				r[epoch][code] = status.counter
			}
		}
	}

	if len(r) == 0 {
		return nil, fmt.Errorf("the given values were out of range: %v to %v", fromEpoch, toEpoch)
	}

	return r, nil
}

func (m *Measurement) addStatusCodeRecord(epoch int64, statusCode int) {
	m.period[epoch].statuses[statusCode] = newStatus()
}

func (m *Measurement) isRecordedAt(epoch int64) bool {
	_, ok := m.period[epoch]
	return ok
}

func (m *Measurement) insertSecondWithLockContext(epoch int64, statusCode int) {
	m.withLockContext(func() {
		// double-check the given key is empty
		if !m.isRecordedAt(epoch) {
			m.period[epoch] = newSecond(statusCode)
			m.addStatusCodeRecord(epoch, statusCode)
		}
	})
}

func (m *Measurement) insertStatusCodeRecordWithLockContext(epoch int64, statusCode int) {
	m.withLockContext(func() {
		if !m.hasStatusCodeRecord(epoch, statusCode) {
			m.addStatusCodeRecord(epoch, statusCode)
		}
	})
}

func (m *Measurement) hasStatusCodeRecord(epoch int64, statusCode int) bool {
	_, ok := m.period[epoch].statuses[statusCode]
	return ok
}

func (m *Measurement) withLockContext(fn func()) {
	m.Lock()
	defer m.Unlock()

	fn()
}
