package retry

import (
	"fmt"
	"io"
	"time"
)

var DefaultRetryer = Retry(Max(10), Timeout(30*time.Second), EOF(), Over(300))

type strategies []Retryer

// We retry (output Yes) if one or more retriers evaluate to Yes and none says No 
// To output Yes some can also evaluate to Maybe as long as one evaluates to Yes. 
// If all evaluate to Maybe we will return Maybe.
// If one evaluates to No we'll return No regardless what the others returned
func (s strategies) Retry(a Attempt) (Decision, error) {
	var errorResult error
	retryResult := Maybe
	for _, try := range s {
		retry, err := try(a);

		// If one evaluates to No we can return immediatly and deliver the related error
		if retry == No {
			return retry, err 
		}

		// Otherwise evaluate further
		retryResult *= retry

		// Resulting error will be last error output by a retrier evaluating to Yes or Maybe
		if err != nil {
	    	errorResult = err
		}
	}

	// Truncate to Yes
	if retryResult > Yes {
		retryResult = Yes
	}

	return retryResult, errorResult
}

func Retry(conditions ...Retryer) Retryer {
	return strategies(conditions).Retry
}

// "Forbidders" (return No or Maybe)

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
			return No, TimeoutError{limit}
		}
		return Maybe, nil
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
			return No, MaxError{limit}
		}
		return Maybe, nil
	}
}

// "Validators" (return Yes or Maybe)

// Errors errors when there is any error
func Errors() Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Err != nil {
			return Yes, a.Err
		}
		return Maybe, nil
	}
}

// EOF retries only when the error is EOF
func EOF() Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Err == io.EOF {
			return Yes, io.EOF
		}
		return Maybe, nil
	}
}

// Over retries when a response is missing or the status code is over a value like 300
func Over(statusCode int) Retryer {
	return func(a Attempt) (Decision, error) {
		if a.Response == nil || a.Response.StatusCode >= statusCode {
			return Yes, nil
		}
		return Maybe, nil
	}
}
