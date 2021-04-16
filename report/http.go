package report

import (
	"net/http"
	"time"
)

// HTTPMiddleware returns a composable factory for HTTP handlers that report the
// result to the given Reporter.
func HTTPMiddleware(reporter Reporter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := &eventRecorder{
				ResponseWriter: w,
				event: Event{
					// Size & Status possiblly overwritten by the ResponseWriter interface
					Status:         200,
					Time:           time.Now().UTC(),
					Method:         r.Method,
					Url:            r.RequestURI,
					Path:           r.URL.Path,
					Proto:          r.Proto,
					Host:           r.Host,
					RemoteAddr:     r.RemoteAddr,
					ForwardedFor:   r.Header.Get("X-Forwarded-For"),
					ForwardedProto: r.Header.Get("X-Forwarded-Proto"),
					Authorization:  r.Header.Get("Authorization"),
					Referrer:       r.Header.Get("Referer"),
					UserAgent:      r.Header.Get("User-Agent"),
					Range:          r.Header.Get("Range"),
					RequestId:      r.Header.Get("X-Request-Id"),
					Region:         r.Header.Get("X-Region"),
					Country:        r.Header.Get("X-Country"),
					City:           r.Header.Get("X-City"),
				},
			}

			start := time.Now()

			next.ServeHTTP(writer, r)

			writer.event.Ms = int(time.Since(start) / time.Millisecond)

			reporter.Report(writer.event)
		})
	}
}
