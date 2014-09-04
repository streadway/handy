package retry

import (
	"fmt"
	"net/http"
	"testing"
)

type errorRoundTrip struct {
	err   error
	count int
}

func (rt *errorRoundTrip) RoundTrip(*http.Request) (*http.Response, error) {
	rt.count++
	return nil, error(rt.err)
}

func TestRetryMaxShouldReturnMaxError(t *testing.T) {
	const attempts = 2

	var (
		req, _ = http.NewRequest("GET", "http://example/test", nil)
		next   = &errorRoundTrip{err: fmt.Errorf("next")}
		trans  = Transport{
			Retry: Max(attempts),
			Next:  next,
		}
	)

	_, err := trans.RoundTrip(req)

	if have, got := next.err.Error(), err.Error(); have == got {
		t.Fatalf("expected to override error from next")
	}

	if want, got := attempts, next.count; want != got {
		t.Fatalf("expected to make %d attempts, got %d", want, got)
	}
}
