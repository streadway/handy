package breaker

import (
	"testing"
	"time"
)

func TestNewCircuitAllows(t *testing.T) {
	c := NewCircuit(0)

	if !c.Allow() {
		t.Fatal("expected new breaker to be closed")
	}
}

func TestBreakerSuccessClosesOpenCircuit(t *testing.T) {
	c := NewCircuit(0)

	c.Trip()

	if c.Allow() {
		t.Fatal("expected new breaker to be open after being tripped")
	}

	c.Success(time.Duration(0))

	if !c.Allow() {
		t.Fatal("expected new breaker to be closed after a success")
	}
}

func TestBreakerFailTripsCircuitWithASingleFailureAt0PrecentThreshold(t *testing.T) {
	c := NewCircuit(0)

	for i := 0; i < 100; i++ {
		c.Success(0)
	}

	c.Failure(0)

	if c.Allow() {
		t.Fatalf("expected failure to not trip circuit at 0%% threshold")
	}
}

func TestBreakerFailDoesNotTripCircuitAt1PrecentThreshold(t *testing.T) {
	const threshold = 0.01

	c := NewCircuit(threshold)

	for i := 0; i < 100-100*threshold; i++ {
		c.Success(0)
	}

	for i := 0; i < 100*threshold; i++ {
		c.Failure(0)
	}

	if !c.Allow() {
		t.Fatalf("expected failure to not trip circuit at 1%% threshold")
	}

	c.Failure(0)

	if c.Allow() {
		t.Fatal("expected failure to trip over the threshold")
	}
}

func TestBreakerAllowsASingleRequestAfterNapTime(t *testing.T) {
	after := make(chan time.Time)

	c := newCircuit(circuitConfig{
		Window: 5 * time.Second,
		After:  func(time.Duration) <-chan time.Time { return after },
	})

	c.Trip()

	after <- time.Now()

	if !c.Allow() {
		t.Fatal("expected to allow once after nap time")
	}

	if c.Allow() {
		t.Fatal("expected to only allow once after nap time")
	}
}

func TestBreakerClosesAfterSuccessAfterNapTime(t *testing.T) {
	after := make(chan time.Time)

	c := newCircuit(circuitConfig{
		Window: 5 * time.Second,
		After:  func(time.Duration) <-chan time.Time { return after },
	})

	c.Trip()

	after <- time.Now()

	if !c.Allow() {
		t.Fatal("expected to allow once after nap time")
	}

	c.Success(time.Duration(0))

	if !c.Allow() {
		t.Fatal("expected to close after first success")
	}

	if !c.Allow() {
		t.Fatal("expected to stay closed after first success")
	}
}
