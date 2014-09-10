package retry

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestComposeOverStatusCode(t *testing.T) {
	retry, err := All(Over(400), Max(2), Timeout(10*time.Second))(Attempt{
		Start:    time.Now(),
		Count:    1,
		Response: &http.Response{StatusCode: 400},
	})
	if err != nil {
		t.Fatalf("expected to not return an error when over status code, but did: %s", err.Error())
	}
	if want, got := Retry, retry; want != got {
		t.Fatalf("expected to retry because over status code, did not")
	}
}

func TestComposeIgnoreOverNilResponse(t *testing.T) {
	retry, err := All(Over(400), Max(2), Timeout(10*time.Second))(Attempt{
		Start:    time.Now(),
		Count:    1,
		Response: nil,
	})

	if err != nil {
		t.Fatalf("expected to not return an error when response is nil, but did: %s", err.Error())
	}
	if want, got := Ignore, retry; want != got {
		t.Fatalf("expected to ignore because response is nil, did not")
	}
}

func TestComposeTimeout(t *testing.T) {
	retry, err := All(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now().Add(-20 * time.Second),
		Count: 1,
	})
	if err == nil {
		t.Fatalf("expected to return an error at timeout, did not")
	}
	if _, isTimeout := err.(TimeoutError); !isTimeout {
		t.Fatalf("expected error to be of type TimeoutError, was not")
	}
	if want, got := Abort, retry; want != got {
		t.Fatalf("expected to timeout, did not")
	}
}

func TestComposeMax(t *testing.T) {
	retry, err := All(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now(),
		Count: 3,
	})
	if err == nil {
		t.Fatalf("expected to return an error at max, did not")
	}
	if _, isMaxError := err.(MaxError); !isMaxError {
		t.Fatalf("expected error to be of type MaxError, was not")
	}
	if want, got := Abort, retry; want != got {
		t.Fatalf("expected to max, did not")
	}
}

func TestComposeErrors(t *testing.T) {
	retry, err := All(Max(2), Timeout(10*time.Second), Errors())(Attempt{
		Start: time.Now(),
		Count: 1,
		Err:   fmt.Errorf("some error"),
	})
	if want, got := Retry, retry; want != got {
		t.Fatalf("expected to retry on error, did not")
	}
	if err != nil {
		t.Fatalf("expected error to be nil when retried")
	}
}

func TestComposeEOF(t *testing.T) {
	retry, err := All(Max(2), Timeout(10*time.Second), EOF())(Attempt{
		Start: time.Now(),
		Count: 1,
		Err:   io.EOF,
	})
	if want, got := Retry, retry; want != got {
		t.Fatalf("expected to retry on EOF error, did not")
	}
	if err != nil {
		t.Fatalf("expected error to be nil when retried")
	}
}

func TestComposeSuccess(t *testing.T) {
	retry, err := All(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now(),
		Count: 1,
	})
	if err != nil {
		t.Fatalf("expected to not return an error when retrying")
	}
	if want, got := Ignore, retry; want != got {
		t.Fatalf("expected to return Ignore, did not")
	}
}

func TestNetDialError(t *testing.T) {
	_, err := net.Dial("tcp", "missing-name:1")
	if err == nil {
		t.Fatalf("expected dial to produce an error to test")
	}
	retry, netErr := Net()(Attempt{
		Start: time.Now(),
		Count: 1,
		Err:   err,
	})

	if want, got := Retry, retry; want != got {
		t.Fatalf("expected Net to %v on dial error, got: %v", want, got)
	}
	if netErr != nil {
		t.Fatalf("expected Net not to return an error, got: %v", netErr)
	}
}
