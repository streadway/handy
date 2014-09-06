package retry

import (
	"testing"
	"time"
	"net/http"
	"fmt"
	"io"
)

func TestComposeOverStatusCode(t *testing.T) {
	response := &http.Response{StatusCode: 400}

	retry, err := Retry(Over(400), Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now(),
		Count: 1,
		Response: response,
	})

	if err != nil {
		t.Fatalf("expected to not return an error when over status code, but did: %s", err.Error())
	}

	if want, got := Yes, retry; want != got {
		t.Fatalf("expected to retry because over status code, did not")
	}
}

func TestComposeOverNilResponse(t *testing.T) {
	response := &http.Response{}
	response = nil

	retry, err := Retry(Over(400), Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now(),
		Count: 1,
		Response: response,
	})

	if err != nil {
		t.Fatalf("expected to not return an error when response is nil, but did: %s", err.Error())
	}

	if want, got := Yes, retry; want != got {
		t.Fatalf("expected to retry because response is nil, did not")
	}
}

func TestComposeTimeout(t *testing.T) {
	retry, err := Retry(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now().Add(-20 * time.Second),
		Count: 1,
	})
	if err == nil {
		t.Fatalf("expected to return an error at timeout, did not")
	}
	_, ok := err.(TimeoutError); if !ok {
		t.Fatalf("expected error to be of type TimeoutError, was not")
	}
	if want, got := No, retry; want != got {
		t.Fatalf("expected to timeout, did not")
	}
}

func TestComposeMax(t *testing.T) {
	retry, err := Retry(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now(),
		Count: 3,
	})
	if err == nil {
		t.Fatalf("expected to return an error at max, did not")
	}
	_, ok := err.(MaxError); if !ok {
		t.Fatalf("expected error to be of type MaxError, was not")
	}
	if want, got := No, retry; want != got {
		t.Fatalf("expected to max, did not")
	}
}

func TestComposeErrors(t *testing.T) {
	innerErr := fmt.Errorf("some error")
	retry, err := Retry(Max(2), Timeout(10*time.Second), Errors())(Attempt{
		Start: time.Now(),
		Count: 1,
		Err: innerErr,
	})


	if err == nil {
		t.Fatalf("expected error to be not nil, but was nil")
	}
	if err != innerErr {
		t.Fatalf("expected to return inner error, did not")
	}
	if want, got := Yes, retry; want != got {
		t.Fatalf("expected to retry on error, did not")
	}
}

func TestComposeEOF(t *testing.T) {
	retry, err := Retry(Max(2), Timeout(10*time.Second), EOF())(Attempt{
		Start: time.Now(),
		Count: 1,
		Err: io.EOF,
	})
	if err == nil {
		t.Fatalf("expected error to be not nil, but was")
	}
	if err != io.EOF {
		t.Fatalf("expected error to be EOF, but was not")
	}	
	if want, got := Yes, retry; want != got {
		t.Fatalf("expected to retry on EOF error, did not")
	}
}

func TestComposeSuccess(t *testing.T) {
	retry, err := Retry(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now(),
		Count: 1,
	})
	if err != nil {
		t.Fatalf("expected to not return an error when retrying")
	}
	if want, got := Maybe, retry; want != got {
		t.Fatalf("expected to return Maybe, did not")
	}
}
