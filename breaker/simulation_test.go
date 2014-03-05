package breaker

import (
	"testing"
	"time"
)

func failrate(c Circuit, count int, pct float64) {
	chance := int(1 / pct)
	if chance <= 0 {
		chance = 1
	}

	for i := 0; i < count; i++ {
		if c.Allow() {
			if (i%count)%chance == 0 {
				c.Failure(0)
			} else {
				c.Success(0)
			}
		}
	}
}

func TestSimulateConcurrentBreakerHandlerWithPartialFailures(t *testing.T) {
	const requestsPerSecond = 100
	const seconds = 5

	now := time.Now()
	after := make(chan time.Time)

	cb := newCircuit(circuitConfig{
		Window:          seconds * time.Second,
		MinObservations: requestsPerSecond / seconds,
		FailureRatio:    0.05,
		Now:             func() time.Time { return now },
		After:           func(time.Duration) <-chan time.Time { return after },
	})

	for i := 0; i < seconds; i++ {
		failrate(cb, requestsPerSecond, 0.20)
		now = now.Add(time.Second)
	}

	if got, want := cb.Allow(), false; got != want {
		t.Fatalf("expected to trip at a high failure rate")
	}

	after <- now

	if got, want := cb.Allow(), true; got != want {
		t.Fatalf("expected to allow in half-open state after cooldown")
	}

	cb.Success(time.Duration(0))

	if got, want := cb.Allow(), true; got != want {
		t.Fatalf("expected to close after success from half-open")
	}

	for i := 0; i < seconds; i++ {
		failrate(cb, requestsPerSecond, 0.02)
		now = now.Add(time.Second)
	}

	if got, want := cb.Allow(), true; got != want {
		t.Fatalf("expected to stay closed after lower error rate")
	}

	for i := 0; i < seconds; i++ {
		failrate(cb, requestsPerSecond, 0.06)
		now = now.Add(time.Second)
	}

	if got, want := cb.Allow(), false; got != want {
		t.Fatalf("expected to open after high error rate again")
	}
}
