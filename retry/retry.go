package retry

import "net/http"

// Attempt counts the round trips issued, starting from 1.  Response is valid
// only when Err is nil.
type Attempt struct {
	Count uint
	Err   error
	*http.Request
	*http.Response
}

// Delayer sleeps or selects any amount of time for each attempt
type Delayer func(Attempt)

// Retryer chooses whether or not to retry this request
type Retryer func(Attempt) bool

type Transport struct {
	// Delay is called for attempts that are retried
	Delay Delayer

	// Retry is called for every attempt
	Retry Retryer

	// Next is called for every attempt
	Next http.RoundTripper
}

// RoundTrip delegates a RoundTrip, then determines via Retry whether to retry
// and Delay for the wait time between attempts.
func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for count, retry := 1, true; retry; count++ {
		res, err := t.Next(req)

		attempt := Attempt{
			Count:    count,
			Err:      err,
			Request:  req,
			Response: res,
		}

		retry = t.Retry(attempt)
		if !retry {
			return res, err
		}

		t.Delay(attempt)
	}
}
