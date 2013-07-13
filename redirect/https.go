package redirect

import (
	"net/http"
)

// HTTPS redirects incoming requests with http.StatusFound that do not have a
// scheme or X-Forwarded-Proto header as 'https' to the same or provided host.
//
// The http.Request.URL.Scheme will modified to 'https' and optionally the
// URL.Host will be replaced with the provided host.
func HTTPS(host string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO handle the Forwarded-For header when it ratifies
		// http://tools.ietf.org/html/draft-ietf-appsawg-http-forwarded-10
		if r.URL.Scheme != "https" && r.Header.Get("X-Forwarded-Proto") != "https" {
			if len(host) > 0 {
				r.URL.Host = host
			}
			r.URL.Scheme = "https"
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
