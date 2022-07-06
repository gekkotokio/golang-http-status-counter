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
func NewMeasurement() Measurement {
	s := newSecond(http.StatusOK)
	t := time.Now().Unix()
	m := map[int64]*second{t: s}

	return Measurement{period: m}
}

// CountUp increases the number of the given HTTP status codes with thread-safe.
func (m *Measurement) CountUp(httpStatus int) {
	epoch := time.Now().Unix()

	if !m.isRecordedAt(epoch) {
		m.insertSecondWithLockContext(epoch, httpStatus)
	}

	m.period[epoch].statuses[httpStatus].incrementWithLockContext()
}

func (m *Measurement) isRecordedAt(epoch int64) bool {
	_, ok := m.period[epoch]
	return ok
}

func (m *Measurement) insertSecondWithLockContext(epoch int64, httpStatus int) {
	m.withLockContext(func() {
		m.period[epoch] = newSecond(httpStatus)
	})
}

func (m *Measurement) withLockContext(fn func()) {
	m.Lock()
	defer m.Unlock()

	fn()
}
