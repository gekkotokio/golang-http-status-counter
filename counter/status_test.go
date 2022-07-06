package counter

import "testing"

func TestIncrement(t *testing.T) {
	s := newStatus()

	if s.counter != 0 {
		t.Errorf("expected value was 0 but %v", s.counter)
	}

	s.increment()

	if s.counter != 1 {
		t.Errorf("expected value was 1 but %v", s.counter)
	}
}

func TestStatusReset(t *testing.T) {
	s := newStatus()

	for i := 0; i < 100; i++ {
		s.increment()
	}

	if s.counter != 100 {
		t.Errorf("expected value was 100 but %v", s.counter)
	}

	s.reset()

	if s.counter != 0 {
		t.Errorf("expected value was 0 but %v", s.counter)
	}
}

func TestIncrementWithLockContext(t *testing.T) {
	s := newStatus()

	if s.counter != 0 {
		t.Errorf("expected value was 0 but %v", s.counter)
	}

	s.incrementWithLockContext()

	if s.counter != 1 {
		t.Errorf("expected value was 1 but %v", s.counter)
	}
}

func TestStatusResetWithLockContext(t *testing.T) {
	s := newStatus()

	for i := 0; i < 100; i++ {
		s.incrementWithLockContext()
	}

	if s.counter != 100 {
		t.Errorf("expected value was 100 but %v", s.counter)
	}

	s.resetWithLockContext()

	if s.counter != 0 {
		t.Errorf("expected value was 0 but %v", s.counter)
	}
}
