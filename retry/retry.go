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

// Retryer chooses whether or not to retry this request, and if not, the error
// to return instead of the prior error.  Composing Retryers with Retry.
type Retryer func(Attempt) (bool, error)

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
		res, err := t.Next.RoundTrip(req)

		attempt := Attempt{
			Start:    start,
			Count:    uint(count),
			Err:      err,
			Request:  req,
			Response: res,
		}

		retry, retryErr := retryer(attempt)

		if retryErr != nil {
			return res, retryErr
		}

		if !retry {
			return res, err
		}

		if t.Delay != nil {
			t.Delay(attempt)
		}
	}
	panic("unreachable")
}
