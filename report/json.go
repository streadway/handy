// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

package report

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// JSONReporter is an event sink that writes serializes into JSON.
type JSONReporter struct {
	mu  sync.Mutex
	out *json.Encoder
}

// NewJSONReporter initializes a new JSONReporter writing to the given
// io.Writer. Concurrent event reports are written out one after another.
func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{
		mu:  sync.Mutex{},
		out: json.NewEncoder(w),
	}
}

// Report implements the Reporter interface.
func (r *JSONReporter) Report(e Event) {
	r.mu.Lock()
	_ = r.out.Encode(e)
	r.mu.Unlock()
}

// JSONMiddleware returns a composable handler factory implementing the JSON
// handler.
func JSONMiddleware(writer io.Writer) func(http.Handler) http.Handler {
	return HTTPMiddleware(NewJSONReporter(writer))
}

// JSON writes a JSON encoded Event to the provided writer at the
// completion of each request.
func JSON(writer io.Writer, next http.Handler) http.Handler {
	return JSONMiddleware(writer)(next)
}
