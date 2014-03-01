package breaker

import (
	"testing"
	"time"
)

func TestNewCircuitBreakerAllows(t *testing.T) {
	b := NewCircuitBreaker(0)

	if !b.Allow() {
		t.Fatal("expected new breaker to be closed")
	}
}

func TestBreakerSuccessClosesOpenCircuit(t *testing.T) {
	b := NewCircuitBreaker(0)

	b.Trip()

	if b.Allow() {
		t.Fatal("expected new breaker to be open after being tripped")
	}

	b.Success(time.Duration(0))

	if !b.Allow() {
		t.Fatal("expected new breaker to be closed after a success")
	}
}

func TestBreakerFailTripsCircuitWithASingleFailureAt0PrecentThreshold(t *testing.T) {
	b := NewCircuitBreaker(0)

	for i := 0; i < 100; i++ {
		b.Success(0)
	}

	b.Failure(0)

	if b.Allow() {
		t.Fatalf("expected failure to not trip circuit at 0%% threshold")
	}
}

func TestBreakerFailDoesNotTripCircuitAt1PrecentThreshold(t *testing.T) {
	const threshold = 0.01

	b := NewCircuitBreaker(threshold)

	for i := 0; i < 100-100*threshold; i++ {
		b.Success(0)
	}

	for i := 0; i < 100*threshold; i++ {
		b.Failure(0)
	}

	if !b.Allow() {
		t.Fatalf("expected failure to not trip circuit at 1%% threshold")
	}

	b.Failure(0)

	if b.Allow() {
		t.Fatal("expected failure to trip over the threshold")
	}
}

func TestBreakerAllowsASingleRequestAfterNapTime(t *testing.T) {
	after := make(chan time.Time)

	b := newCircuitBreaker(circuitBreakerConfig{
		Window: 5 * time.Second,
		After:  func(time.Duration) <-chan time.Time { return after },
	})

	b.Trip()

	after <- time.Now()

	if !b.Allow() {
		t.Fatal("expected to allow once after nap time")
	}

	if b.Allow() {
		t.Fatal("expected to only allow once after nap time")
	}
}

func TestBreakerClosesAfterSuccessAfterNapTime(t *testing.T) {
	after := make(chan time.Time)

	b := newCircuitBreaker(circuitBreakerConfig{
		Window: 5 * time.Second,
		After:  func(time.Duration) <-chan time.Time { return after },
	})

	b.Trip()

	after <- time.Now()

	if !b.Allow() {
		t.Fatal("expected to allow once after nap time")
	}

	b.Success(time.Duration(0))

	if !b.Allow() {
		t.Fatal("expected to close after first success")
	}

	if !b.Allow() {
		t.Fatal("expected to stay closed after first success")
	}
}
