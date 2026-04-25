package middlewares

// Package mw provides composable HTTP middleware primitives and semantic helpers.
import (
	"net/http"

	fun "github.com/MintzyG/FastUtilitiesNet"
)

// QueryRequire rejects requests where any of the listed query params is absent or empty.
func QueryRequire(params ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			for _, p := range params {
				if !q.Has(p) || q.Get(p) == "" {
					fun.BadRequest("missing required query param: " + p).Send(w)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// QueryAllow rejects requests that contain any query param not in the allowed list.
func QueryAllow(allowed ...string) func(http.Handler) http.Handler {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for param := range r.URL.Query() {
				if _, ok := set[param]; !ok {
					fun.BadRequest("unknown query param: " + param).Send(w)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// QueryDefault sets param to value if it is absent or empty.
func QueryDefault(param, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get(param) == "" {
				q.Set(param, value)
				r.URL.RawQuery = q.Encode()
			}
			next.ServeHTTP(w, r)
		})
	}
}
