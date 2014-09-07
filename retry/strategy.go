package retry

import (
	"fmt"
	"io"
	"time"
)

var DefaultRetryer = RetryBy(Max(10), Timeout(30*time.Second), EOF(), Over(300))

type strategies []Retryer

// We retry (output Retry) if one or more retriers evaluate to Retry and none says Abort
// To output Retry some can also evaluate to Ignore as long as one evaluates to Retry.
// If all evaluate to Ignore we will return Ignore (this should be output for a successful attempt whose response will be returned as is and we may stop retrying).
// If one evaluates to Abort we'll return Abort regardless what the others returned
func (s strategies) RetryBy(a Attempt) (Decision, error) {
	var errorResult error
	retryResult := Ignore
	for _, try := range s {
		retry, err := try(a)

		// If one evaluates to No we can return immediatly and deliver the related error
		if retry == Abort {
			return retry, err
		}

		// Otherwise evaluate further
		retryResult *= retry

		// Resulting error will be last error output by a retrier evaluating to Retry or Ignore
		if err != nil {
			errorResult = err
		}
	}

	// Truncate to Retry
	if retryResult > Retry {
		retryResult = Retry
	}

	return retryResult, errorResult
}

func RetryBy(conditions ...Retryer) Retryer {
	return strategies(conditions).RetryBy
}

// "Forbidders" (return Abort or Ignore)

type TimeoutError struct {
	Duration time.Duration
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("timed out after %.2f seconds", e.Duration.Seconds())
}

// Timeout errors after a duration of time passes since the first attempt.
func Timeout(limit time.Duration) Retryer {
	return func(a Attempt) (Decision, error) {
		if time.Since(a.Start) >= limit {
			return Abort, TimeoutError{limit}
		}
		return Ignore, nil
	}
}

type MaxError struct {
	Attempts uint
}

func (e MaxError) Error() string {
	return fmt.Sprintf("retry limit exceeded after %d attempts", e.Attempts)
}

// Max errors after a limited number of attempts
func Max(limit uint) Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Count >= limit {
			return Abort, MaxError{limit}
		}
		return Ignore, nil
	}
}

// "Validators" (return Retry or Ignore)

// Errors errors when there is any error
func Errors() Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Err != nil {
			return Retry, a.Err
		}
		return Ignore, nil
	}
}

// EOF retries only when the error is EOF
func EOF() Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Err == io.EOF {
			return Retry, io.EOF
		}
		return Ignore, nil
	}
}

// Over retries when a response is missing or the status code is over a value like 300
func Over(statusCode int) Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Response == nil || a.Response.StatusCode >= statusCode {
			return Retry, nil
		}
		return Ignore, nil
	}
}
