package rewrite

import (
	"net/http"
)

// Method modifies the http.Request.Method for POST requests to the form value
// "_method" only if that value is one of PUT, PATCH or DELETE.
func Method(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			switch _method := r.FormValue("_method"); _method {
			case "PUT", "PATCH", "DELETE":
				r.Method = _method
			}
		}
		next.ServeHTTP(w, r)
	})
}
