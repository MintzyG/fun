package middlewares

// Provides structured HTTP request logging middleware.
import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Config controls logger behavior.
type Config struct {
	// Logger is the zap logger to write to. Required.
	Logger *zap.Logger

	// SkipPrefixes lists URL path prefixes that will not be logged.
	// e.g. []string{"/admin/asynq", "/healthz"}
	SkipPrefixes []string

	// RequestIDHeader is the header name to read the request ID from.
	// Defaults to "X-Request-ID".
	RequestIDHeader string
}

func (c *Config) requestIDHeader() string {
	if c.RequestIDHeader != "" {
		return c.RequestIDHeader
	}
	return "X-Request-ID"
}

// Logs returns structured request logging middleware.
// It reads the request ID from the header set by the RequestID middleware (or any upstream proxy).
func Logs(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			skipRoute(w, r, next, cfg.SkipPrefixes)

			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			routePattern := resolveRoutePattern(r)
			reqID := r.Header.Get(cfg.requestIDHeader())

			cfg.Logger.Info("http_request",
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", routePattern),
				zap.Int("status", ww.status),
				zap.Duration("duration", time.Since(start)),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}
