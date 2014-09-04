package retry

import (
	"fmt"
	"io"
	"time"
)

var DefaultRetryer = Retry(Max(10), Timeout(30*time.Second), EOF(), Over(300))

type strategies []Retryer

func (s strategies) Retry(a Attempt) (bool, error) {
	for _, try := range s {
		if retry, err := try(a); !retry || err != nil {
			return retry, err
		}
	}
	return true, nil
}

// Retry stops retrying when any retryer returns false or an error.
func Retry(conditions ...Retryer) Retryer {
	return strategies(conditions).Retry
}

// Over retries when a response is missing or the status code is over a value like 300
func Over(statusCode int) Retryer {
	return func(a Attempt) (bool, error) {
		return a.Response == nil || a.Response.StatusCode >= statusCode, nil
	}
}

// Timeout errors after a duration of time passes since the first attempt.
func Timeout(limit time.Duration) Retryer {
	return func(a Attempt) (bool, error) {
		if time.Since(a.Start) >= limit {
			return false, fmt.Errorf("timeout limit %s exceeded", limit)
		}
		return true, nil
	}
}

// Max errors after a limited number of attempts
func Max(limit uint) Retryer {
	return func(a Attempt) (bool, error) {
		if a.Count >= limit {
			return false, fmt.Errorf("retry limit %d exceeded", limit)
		}
		return true, nil
	}
}

// Errors errors when there is any error
func Errors() Retryer {
	return func(a Attempt) (bool, error) {
		return a.Err != nil, a.Err
	}
}

// EOF retries when there is no error or that error is EOF
func EOF() Retryer {
	return func(a Attempt) (bool, error) {
		return a.Err == nil || a.Err == io.EOF, nil
	}
}
