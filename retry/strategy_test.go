package retry

import (
	"testing"
	"time"
)

func TestComposeTimeout(t *testing.T) {
	retry, err := Retry(Max(2), Timeout(10*time.Second))(Attempt{
		Start: time.Now().Add(-20 * time.Second),
		Count: 1,
	})
	if err == nil {
		t.Fatalf("expected to return an error at timeout, did not")
	}
	if want, got := false, retry; want != got {
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
	if want, got := false, retry; want != got {
		t.Fatalf("expected to max, did not")
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
	if want, got := true, retry; want != got {
		t.Fatalf("expected to retry, did not")
	}
}
