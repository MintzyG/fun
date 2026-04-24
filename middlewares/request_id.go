package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const requestIDKey ctxKey = "middleware.requestID"

// RequestIDConfig controls RequestID middleware behavior.
type RequestIDConfig struct {
	// Header is the header to read/write the request ID.
	// Defaults to "X-Request-ID".
	Header string
}

func (c RequestIDConfig) header() string {
	if c.Header != "" {
		return c.Header
	}
	return "X-Request-ID"
}

// RequestID adds a request ID to the context and response headers.
// If the incoming request already carries one in the configured header it is reused;
// otherwise a new UUID v7 is generated (falling back to v4 on error).
func RequestID(cfg ...RequestIDConfig) func(http.Handler) http.Handler {
	var c RequestIDConfig
	if len(cfg) > 0 {
		c = cfg[0]
	}
	header := c.header()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			reqID := r.Header.Get(header)
			if reqID == "" {
				if uid, err := uuid.NewV7(); err == nil {
					reqID = uid.String()
				} else {
					reqID = uuid.New().String()
				}
			}

			ctx = context.WithValue(ctx, requestIDKey, reqID)
			w.Header().Set(header, reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromCtx retrieves the request ID stored by the RequestID middleware.
func RequestIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}
