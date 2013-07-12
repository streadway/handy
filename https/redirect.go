package https

import (
	"net/http"
)

// All http requests are getting redirected to https url.
// To detect a https connections the url scheme and the header "X-Forwarded-Proto" are checked.
//
// The host parameter is used to build the redirect url
func Redirect(host string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Scheme != "https" && r.Header.Get("X-Forwarded-Proto") != "https" {
			r.URL.Host = host
			r.URL.Scheme = "https"
			http.Redirect(w, r, r.URL.String(), 302)
			return
		}

		next.ServeHTTP(w, r)
	})
}
