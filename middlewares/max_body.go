package middlewares

// Package bodymw provides middleware for limiting HTTP request body size.
import "net/http"

// MaxBodySize limits the request body to limit bytes.
// Requests with bodies exceeding the limit will fail when the handler reads past it.
// Only applied when the request body is non-nil.
func MaxBodySize(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, limit)
			}
			next.ServeHTTP(w, r)
		})
	}
}
