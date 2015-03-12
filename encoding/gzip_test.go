// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the README file.
// Source code and contact info at http://github.com/streadway/handy

package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type plain string

func (h plain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h))
}

type json string

func (h json) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=utf8")
	w.Write([]byte(h))
}

type file struct {
	name string
	data []byte
}

func (h file) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, h.name, time.Time{}, bytes.NewReader(h.data))
}

func decode(t *testing.T, body io.Reader) string {
	plain := &bytes.Buffer{}
	gz, err := gzip.NewReader(body)
	if err != nil {
		t.Fatalf("expected a gzip stream, got: %q", err)
	}
	io.Copy(plain, gz)
	return plain.String()
}

func acceptGzip() *http.Request {
	return &http.Request{
		Header: http.Header{"Accept-Encoding": {"gzip"}},
	}
}

func TestGzip(t *testing.T) {
	const msg = "the meaning of life, the universe and everything"

	var (
		handler = Gzip(plain(msg))
		resp    = httptest.NewRecorder()
		req     = acceptGzip()
	)

	handler.ServeHTTP(resp, req)

	if hdr := resp.HeaderMap.Get("Content-Encoding"); hdr != "gzip" {
		t.Fatalf("expected content encoding to be gzip, got: %q", hdr)
	}

	if hdr := resp.HeaderMap.Get("Vary"); hdr != "Accept-Encoding" {
		t.Fatalf("expected to vary on accept encoding, got: %q", hdr)
	}

	if want, got := msg, decode(t, resp.Body); want != got {
		t.Fatalf("expected to decompress message, got: %q", got)
	}
}

func TestGzipWithFile(t *testing.T) {
	var data = `function foo(){return "bar";}`

	handler := Gzip(file{
		name: "foobar.js",
		data: []byte(data),
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	u, _ := url.Parse(server.URL)
	req := &http.Request{
		URL:    u,
		Header: http.Header{"Accept-Encoding": {"gzip"}},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if hdr := resp.Header.Get("Content-Encoding"); hdr != "gzip" {
		t.Fatalf("expected content encoding to be gzip, got: %q", hdr)
	}

	if hdr := resp.Header.Get("Vary"); hdr != "Accept-Encoding" {
		t.Fatalf("expected to vary on accept encoding, got: %q", hdr)
	}

	if want, got := data, decode(t, resp.Body); want != got {
		t.Fatalf("expected to decompress message, got: %q", got)
	}
}

func TestMatchingGzipTypes(t *testing.T) {
	const msg = `{"meaning": 42}`

	var (
		types   = []string{"application/json"}
		handler = GzipTypes(types, json(msg))
		resp    = httptest.NewRecorder()
		req     = acceptGzip()
	)

	handler.ServeHTTP(resp, req)

	if want, got := "gzip", resp.HeaderMap.Get("Content-Encoding"); want != got {
		t.Fatalf("expected content encoding %q, got: %q", want, got)
	}

	if want, got := msg, decode(t, resp.Body); want != got {
		t.Fatalf("expected decoded json stream %q, got: %q", want, got)
	}
}

func TestNonMatchingGzipTypes(t *testing.T) {
	const msg = `just some plain text`

	var (
		types   = []string{"application/json"}
		handler = GzipTypes(types, plain(msg))
		resp    = httptest.NewRecorder()
		req     = acceptGzip()
	)

	handler.ServeHTTP(resp, req)

	if want, got := "", resp.HeaderMap.Get("Content-Encoding"); want != got {
		t.Fatalf("expected no content encoding, got: %q", got)
	}
}
