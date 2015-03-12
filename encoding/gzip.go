// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

/*
Package encoding contains Content-Encoding related filters.
*/
package encoding

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

type gzipWriter struct {
	http.ResponseWriter
	io.WriteCloser
	sync.Mutex
	wroteHeader bool
	types       []string
}

func (w gzipWriter) canGzip() bool {
	if len(w.types) == 0 {
		return true
	}

	contentType := w.Header().Get("Content-Type")
	for _, mediaType := range w.types {
		if strings.Contains(contentType, mediaType) {
			return true
		}
	}

	return false
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	w.Lock()
	defer w.Unlock()

	if !w.wroteHeader {
		w.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Lock()
	}

	if w.WriteCloser == nil {
		if hdr := w.Header().Get("Content-Encoding"); hdr == "gzip" {
			w.WriteCloser = gzip.NewWriter(w.ResponseWriter)
		} else {
			w.WriteCloser = nopCloser{w.ResponseWriter}
		}
	}

	return w.WriteCloser.Write(b)
}

func (w *gzipWriter) WriteHeader(code int) {
	w.Lock()
	defer w.Unlock()

	if !w.wroteHeader && w.canGzip() {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
	}
	w.wroteHeader = true

	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipWriter) Close() error {
	w.Lock()
	defer w.Unlock()

	if w.WriteCloser == nil {
		return nil
	}
	return w.WriteCloser.Close()
}

// Gzip calls the next handler with a response writer that will compress the
// outbound writes.  This filter assumes a chunked transfer encoding, so do not
// add a Content-Length header in the terminal handler.
//
// If the request does not accept a gzip encoding, this filter has no effect.
func Gzip(next http.Handler) http.Handler {
	return GzipTypes(nil, next)
}

// GzipTypes sets the gzips the response if the the request Accept-Encoding
// contains 'gzip' and the response 'Content-Type' contains one of the mediaTypes.
// When no or nil mediaTypes are provided, all content types will be gzip encoded.
func GzipTypes(mediaTypes []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		zipper := &gzipWriter{types: mediaTypes, ResponseWriter: w}
		defer zipper.Close()
		next.ServeHTTP(zipper, r)
	})
}
