package redirect

import (
	"net/http"
)

// HTTPS ensures the scheme of incoming requests is https:// either from the
// http.Request.URL.Scheme or X-Forwarded-Proto header.  When the scheme is not
// https, a redirect with 302 will be made to the same host found in
// http.Request.Host.
func HTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO handle the Forwarded-For header when it ratifies
		// http://tools.ietf.org/html/draft-ietf-appsawg-http-forwarded-10
		if r.URL.Scheme != "https" && r.Header.Get("X-Forwarded-Proto") != "https" {
			r.URL.Scheme = "https"
			r.URL.Host = r.Host
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
