package counter

import (
	"net/http"
	"sync"
	"time"
)

type Measurement struct {
	sync.Mutex
	period map[int64]*second
}

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
