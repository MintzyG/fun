package middlewares

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func resolveRoutePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if p := rctx.RoutePattern(); p != "" {
			return p
		}
	}
	return "not_found"
}

func skipRoute(w http.ResponseWriter, r *http.Request, next http.Handler, skipPrefixes []string) {
	for _, prefix := range skipPrefixes {
		if len(r.URL.Path) >= len(prefix) && r.URL.Path[:len(prefix)] == prefix {
			next.ServeHTTP(w, r)
			return
		}
	}
}

// chain composes multiple middleware into one, applied left to right.
func chain(mws ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// statusWriter captures the status code written by a handler and delegates
// Hijack so WebSocket / SSE handlers work correctly.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("fun: underlying ResponseWriter does not support hijacking")
	}
	return h.Hijack()
}
