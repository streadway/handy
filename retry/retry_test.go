package retry

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testHandler struct {
	statusOne int
	statusTwo int
	threshold int
	counter   *int
	t         *testing.T
}

func (h testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if *h.counter >= h.threshold {
		w.WriteHeader(h.statusTwo)
	} else {
		w.WriteHeader(h.statusOne)
	}
	*h.counter++
	h.t.Logf("Requests counted: %d", *h.counter)
}

/* Generic test with 1 second time-base for delay generator */
func testDefaultTransportAccept3xxsGeneric(t *testing.T, retries, expectedDuration, expectedStatus, goodStatus, badStatus int, strategy Strategy) *Transport {
	counter := 0
	expectedCount := retries + 1
	handler := testHandler{badStatus, goodStatus, retries, &counter, t}
	server := httptest.NewServer(handler)
	defer server.Close()

	transport := newDefaultTransportAccept3xxs(uint(retries), 10*time.Second, 1*time.Second, strategy, func(e Event, err error) {
		t.Logf("Event type %d, message: %s", e, err.Error())
	})

	retryClient := http.Client{
		Transport: transport,
	}

	t.Logf("Trying request ...")

	// Measure request start time
	start := time.Now()

	resp, err := retryClient.Get(server.URL)

	duration := int(math.Floor(time.Since(start).Seconds()))

	if err != nil {
		t.Logf("Got resulting error message (no HTTP response): %s", err.Error())
	} else {
		t.Logf("Got HTTP response, performing validation checks ...")

		if resp.StatusCode != expectedStatus {
			t.Errorf("Response status code is %d but is expected to be %d", resp.StatusCode, expectedStatus)
		}

		if counter != expectedCount {
			t.Errorf("Should have counted %d requests but counted %d", expectedCount, counter)
		}

		if duration != expectedDuration {
			t.Errorf("Total transaction duration should be at least %d seconds but measured %d seconds", expectedDuration, duration)
		}
	}

	return transport
}

func TestDefaultTransportAccept3xxsFibonacci(t *testing.T) {
	// The combination of input values is critical to pass the test
	testDefaultTransportAccept3xxsGeneric(t, 5, 12, 200, 200, 400, Fibonacci)
}

func TestDefaultTransportAccept3xxsConstant(t *testing.T) {
	// The combination of input values is critical to pass the test
	testDefaultTransportAccept3xxsGeneric(t, 5, 5, 200, 200, 400, Constant)
}

func TestTimeout(t *testing.T) {
	// The combination of input values is critical to pass the test
	testDefaultTransportAccept3xxsGeneric(t, 11, 10, 400, 200, 400, Constant)
}

func TestMaxRetries(t *testing.T) {
	// The combination of input values is critical to pass the test
	testDefaultTransportAccept3xxsGeneric(t, 2, 2, 400, 400, 400, Constant)
}
