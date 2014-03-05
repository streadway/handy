package breaker

import (
	"errors"
	"net/http"
	"time"
)

var (
	// ErrCircuitOpen is returned by the transport when the downstream is
	// unavailable due to a broken circuit.
	ErrCircuitOpen = errors.New("circuit open")
)

// Transport is an experimental implementation of a circuit-breaking
// http.RoundTripper that returns ErrCircuitOpen after a 5% failure rate over
// a sliding window of 5 seconds with a 1 second cooldown period before
// retrying with a single request. Failure is defined by the user-provided
// failure function.
func Transport(failure func(*http.Response) bool, next http.RoundTripper) http.RoundTripper {
	return &transport{
		failure: failure,
		circuit: NewCircuit(0.05),
		next:    next,
	}
}

type transport struct {
	failure func(*http.Response) bool
	circuit Circuit
	next    http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.circuit.Allow() {
		return nil, ErrCircuitOpen
	}

	begin := time.Now()
	resp, err := t.next.RoundTrip(req)

	duration := time.Since(begin)
	if err != nil || t.failure(resp) {
		t.circuit.Failure(duration)
	} else {
		t.circuit.Success(duration)
	}

	// TODO metrics
	return resp, err
}

// DefaultFailure reports any response status code greater than or equal to
// 400 as a failure.
func DefaultFailure(resp *http.Response) bool {
	return resp.StatusCode >= 400
}
