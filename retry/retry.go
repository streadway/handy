// Package retry implements a retrying transport based on a combination of strategies.
package retry

import (
	"net/http"
	"time"
)

var now = time.Now

// Attempt counts the round trips issued, starting from 1.  Response is valid
// only when Err is nil.
type Attempt struct {
	Start time.Time
	Count uint
	Err   error
	*http.Request
	*http.Response
}

// Delayer sleeps or selects any amount of time for each attempt.
type Delayer func(Attempt)

type Decision int

const (
	Retry  Decision = 2
	Ignore Decision = 1
	Abort  Decision = 0
)

// Retryer chooses whether or not to retry this request, and if not, the error
// to return instead of the prior error.  Composing Retryers with Retry.
type Retryer func(Attempt) (Decision, error)

type Transport struct {
	// Delay is called for attempts that are retried.  If nil, no delay will be used.
	Delay Delayer

	// Retry is called for every attempt
	Retry Retryer

	// Next is called for every attempt
	Next http.RoundTripper
}

// RoundTrip delegates a RoundTrip, then determines via Retry whether to retry
// and Delay for the wait time between attempts.
func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		retryer = t.Retry
		start   = now()
	)
	if retryer == nil {
		retryer = DefaultRetryer
	}

	for count, retry := 1, true; retry; count++ {
		// Perform request
		res, err := t.Next.RoundTrip(req)

		// Collect result of attempt
		attempt := Attempt{
			Start:    start,
			Count:    uint(count),
			Err:      err,
			Request:  req,
			Response: res,
		}

		// Evaluate attempt
		retry, retryErr := retryer(attempt)

		// Override error by retrier evaluation
		if retryErr != nil {
			err = retryErr
		}

		// Return response and error if we did not evaluate to Retry decision
		if retry != Retry {
			return res, err
		}

		// Delay next attempt
		if t.Delay != nil {
			t.Delay(attempt)
		}
	}
	panic("unreachable")
}
