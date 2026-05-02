package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/MintzyG/fun"
	"go.uber.org/zap"
)

// Recover catches panics, logs them via zap with the full stack trace,
// and responds with 500 Internal Server Error.
//
//	r.Use(mw.Recover(logger))
func Recover(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error("panic recovered",
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Any("panic", rec),
						zap.String("stack", string(stack)),
						zap.String("remote_addr", r.RemoteAddr),
					)
					fun.InternalServerError("internal server error").Send(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
