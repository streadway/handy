package method

import (
	"net/http"
)

// Provides support for PUT, PATCH and DELETE requests by using a POST request together with
// _method parameter which includes the wanted method.
// This handler will override r.Method in that case.
func Override(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			override := r.Header.Get("_method")
			switch override {
			case "PUT", "PATCH", "DELETE":
				r.Method = override
			}
		}
		next.ServeHTTP(w, r)
	})
}
