// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

package report

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJSON(t *testing.T) {
	const worktime = 10 * time.Millisecond
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	out := &bytes.Buffer{}

	h := JSON(out, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(worktime)
	}))

	h.ServeHTTP(res, req)

	report := map[string]interface{}{}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("expected to decode json report, got: %q", err)
	}

	t.Log(report)

	for field, want := range map[string]interface{}{
		"status": float64(200),
		"method": "GET",
		"proto":  "HTTP/1.1",
		"time":   "",
		"ms":     "",
	} {
		if _, ok := report[field]; !ok {
			t.Fatalf("expected to report %q with any value, did not", field)
		}

		if want != "" {
			if got := report[field]; got != want {
				t.Fatalf("expected to report %q with %v, got %v", field, want, got)
			}
		}
	}

	if ms, ok := report["ms"].(float64); !ok {
		t.Fatalf("ms is not a number")
	} else {
		if want, got, delta := worktime, time.Duration(ms)*time.Millisecond, time.Millisecond; want+delta < got || want-delta > got {
			t.Fatalf("duration falls outside of %s±%s, got: %d", want, delta, got)
		}
	}
}
