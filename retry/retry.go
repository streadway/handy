// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

/*
Package retry contains a retrying HTTP transport.
*/
package retry

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type Strategy int
type Event int

type EventHandler func(Event, error)
type DelayFunc func(uint) time.Duration

const (
	Timeout Event = iota
	LimitReached
	Success
	BadHTTPStatus
	EOFRetry
)

const (
	Linear Strategy = iota
	Exponential
	Fibonacci
	Custom
	Constant
)

type Transport struct {
	// Minimum HTTP status to satisfy the roundtrip
	MinHTTPStatus uint

	// Maximum HTTP status to satisfy the roundtrip (exclusive limit)
	MaxHTTPStatus uint

	// Max number of retries
	MaxRetries uint

	// Should we retry on EOF (when remote socket closed)
	RetryOnEOF bool

	// Response time-out
	ResponseTimeout time.Duration

	// Delay time-constant
	DelayTimebase time.Duration

	// Type of retry strategy
	Strategy Strategy

	// Next is the http.RoundTripper on which requests are retried.
	Next http.RoundTripper

	// Optional callback
	EventHandler EventHandler

	// Optional custom delay function
	DelayFunc DelayFunc

	// Possibiity to read a copy of the last response (even if not satisfying desired criterias)
	lastResponse http.Response
}

func newDefaultTransportAccept3xxs(maxRetries uint, responseTimeout time.Duration, delayTimebase time.Duration, strategy Strategy, eventHandler EventHandler) *Transport {
	return &Transport{200,
		400,
		maxRetries,
		true,
		responseTimeout,
		delayTimebase,
		strategy,
		http.DefaultTransport,
		eventHandler,
		nil,
		http.Response{}}
}

func newDefaultTransportReject3xxs(maxRetries uint, responseTimeout, delayTimebase time.Duration, strategy Strategy, eventHandler EventHandler) *Transport {
	return &Transport{200,
		300,
		maxRetries,
		true,
		responseTimeout,
		delayTimebase,
		strategy,
		http.DefaultTransport,
		eventHandler,
		nil,
		http.Response{}}
}

func newDefaultTransportOnly3xxs(maxRetries uint, responseTimeout, delayTimebase time.Duration, strategy Strategy, eventHandler EventHandler) *Transport {
	return &Transport{300,
		400,
		maxRetries,
		true,
		responseTimeout,
		delayTimebase,
		strategy,
		http.DefaultTransport,
		eventHandler,
		nil,
		http.Response{}}
}

func newDefaultTransportAccept4xxs(maxRetries uint, responseTimeout, delayTimebase time.Duration, strategy Strategy, eventHandler EventHandler) *Transport {
	return &Transport{200,
		500,
		maxRetries,
		true,
		responseTimeout,
		delayTimebase,
		strategy,
		http.DefaultTransport,
		eventHandler,
		nil,
		http.Response{}}
}

// Retries error responses with small number of retries
func (t Transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	// Current iteration
	var n uint
	// Measure request start time
	start := time.Now()
	// Perform request(s)
	for {
		res, err = t.Next.RoundTrip(req)

		// Return non-HTTP errors immediately,
		// except EOF that might happen if remote socket force-closes connection
		// which we consider as a retriable server-side failure
		if err != nil && err != io.EOF || err == io.EOF && !t.RetryOnEOF {
			return
		} else if err == io.EOF {
			t.NotifyHandler(EOFRetry, fmt.Errorf("Retrying on EOF, retrial %d of %d", n, t.MaxRetries))
		}

		// If err is nil we can assume response is not nil (store copy of response anyway if client wants to analyze further)
		t.lastResponse = *res

		// Happy path. :)
		if res.StatusCode >= int(t.MinHTTPStatus) && res.StatusCode < int(t.MaxHTTPStatus) {
			t.NotifyHandler(Success, fmt.Errorf("Request succeeded"))
			return
		} else {
			t.NotifyHandler(BadHTTPStatus, fmt.Errorf("Bad HTTP response status (%d), retrial %d of %d", res.StatusCode, n, t.MaxRetries))
		}

		// Somewhat obey the default response timeout by checking we haven't
		// exceeded it including retries.
		if time.Since(start) > t.ResponseTimeout {
			err = fmt.Errorf("Exceeded response timeout (%f seconds)", t.ResponseTimeout.Seconds())
			t.NotifyHandler(Timeout, err)
			return nil, err
		}

		// Check and eventually break loop before doing any sleep
		if n >= t.MaxRetries {
			err = fmt.Errorf("Max retries reached (%d)", t.MaxRetries)
			t.NotifyHandler(LimitReached, err)
			break
		}

		// Apply sleep function with current delay value
		n++
		time.Sleep(t.Delay(n))
	}

	return nil, err
}

func (t *Transport) NotifyHandler(e Event, err error) {
	if t.EventHandler != nil {
		t.EventHandler(e, err)
	}
}

// Default DelayFunc implementation
func (t Transport) Delay(iteration uint) time.Duration {
	switch t.Strategy {
	case Constant:
		return t.DelayTimebase
	case Linear:
		return t.DelayTimebase * time.Duration(iteration)
	case Exponential:
		return time.Duration(float64(t.DelayTimebase) * math.Exp(float64(iteration)))
	case Fibonacci:
		return t.DelayTimebase * time.Duration(fib(int(iteration)))
	case Custom:
		if t.DelayFunc != nil {
			return t.DelayFunc(iteration)
		}
	}
	return 0
}

func fib(i int) int {
	var n0, n1, j int
	for n0, n1, j = 0, 1, 0; j < i; j++ {
		n0, n1 = n1, n0+n1
	}
	return n0
}
