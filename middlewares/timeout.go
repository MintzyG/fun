package middlewares

import (
	"context"
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// Timeout sets a deadline on the request context.
// Handlers must respect ctx.Done() to honour the timeout — this is especially
// important for SSE and WebSocket handlers which own their own read/write loops.
// Pass 0 to use the default of 30s.
//
//	r.Use(mw.Timeout(10 * time.Second))
//	r.Use(mw.Timeout(0)) // 30s default
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	if d <= 0 {
		d = defaultTimeout
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
