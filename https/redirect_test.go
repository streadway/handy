package https

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type code int

func (h code) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(int(h))
}

func TestServesUsingXForwardedProtoHttps(t *testing.T) {
	h := Redirect("localhost", code(204))

	resp := httptest.NewRecorder()
	req := &http.Request{}

	req.Method = "GET"
	req.URL = &url.URL{Scheme: "http", Path: "/"}
	req.Header = map[string][]string{"X-Forwarded-Proto": {"https"}}

	h.ServeHTTP(resp, req)

	if resp.Code != 204 {
		t.Fatalf("expected 200 but got %d", resp.Code)
	}
}

func TestServesUsingHttps(t *testing.T) {
	h := Redirect("localhost", code(204))

	resp := httptest.NewRecorder()
	req := &http.Request{}

	req.Method = "GET"
	req.URL = &url.URL{Scheme: "https", Path: "/index.html"}
	req.Header = map[string][]string{}

	h.ServeHTTP(resp, req)

	if resp.Code != 204 {
		t.Fatalf("expected 200 but got %d", resp.Code)
	}
}

func TestRedirectsUsingHttp(t *testing.T) {
	h := Redirect("localhost", code(204))

	resp := httptest.NewRecorder()
	req := &http.Request{}

	req.Method = "GET"
	req.URL = &url.URL{Scheme: "http", Path: "/index.html"}
	req.Header = map[string][]string{}

	h.ServeHTTP(resp, req)

	if resp.Code != 302 {
		t.Fatalf("expected 302 but got %d", resp.Code)
	}
	if loc := resp.Header().Get("Location"); loc != "https://localhost/index.html" {
		t.Fatalf("location is wrong: %s", loc)
	}
}

func TestRedirectsUsingXForwardedProtoHttp(t *testing.T) {
	h := Redirect("localhost", code(204))

	resp := httptest.NewRecorder()
	req := &http.Request{}

	req.Method = "GET"
	req.URL = &url.URL{Scheme: "http", Path: "/index.html"}
	req.Header = map[string][]string{"X-Forwarded-Proto": {"http"}}

	h.ServeHTTP(resp, req)

	if resp.Code != 302 {
		t.Fatalf("expected 302 but got %d", resp.Code)
	}
	if loc := resp.Header().Get("Location"); loc != "https://localhost/index.html" {
		t.Fatalf("location is wrong: %s", loc)
	}
}
