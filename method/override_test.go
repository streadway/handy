package method

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type expect struct {
	Method string
	t      *testing.T
}

func (h expect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Method != r.Method {
		h.t.Fatalf("expected http method %s but got %s", h.Method, r.Method)
	}
	w.WriteHeader(204)
}

func TestDoesntOverrideGET(t *testing.T) {

	h := Override(expect{Method: "GET", t: t})

	req := &http.Request{
		Method: "GET",
		Header: map[string][]string{
			"_method": {"DELETE"}},
	}

	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestOverridesPOSTWithDELETE(t *testing.T) {

	h := Override(expect{Method: "DELETE", t: t})

	req := &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"_method": {"DELETE"}},
	}

	h.ServeHTTP(httptest.NewRecorder(), req)
}

func TestOverridesPOSTWithPUT(t *testing.T) {

	h := Override(expect{Method: "PUT", t: t})

	req := &http.Request{
		Method: "POST",
		Header: map[string][]string{
			"_method": {"PUT"}},
	}

	h.ServeHTTP(httptest.NewRecorder(), req)
}
