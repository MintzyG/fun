package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

// NoCache sets headers that instruct clients and proxies never to cache the response.
// Use on endpoints that always return fresh data (auth, user-specific, etc).
//
//	r.With(mw.NoCache()).Get("/me", handler)
func NoCache() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Cache-Control", "no-store, no-cache, must-revalidate")
			h.Set("Pragma", "no-cache")
			h.Set("Expires", "0")
			next.ServeHTTP(w, r)
		})
	}
}

// CachePublic sets Cache-Control to public with the given max-age.
// Use on responses that are safe to cache by anyone (CDN, browser, proxy).
//
//	r.With(mw.CachePublic(5 * time.Minute)).Get("/assets/config", handler)
func CachePublic(maxAge time.Duration) func(http.Handler) http.Handler {
	v := fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds()))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", v)
			next.ServeHTTP(w, r)
		})
	}
}

// CachePrivate sets Cache-Control to private with the given max-age.
// Use on user-specific responses that the browser may cache but proxies must not.
//
//	r.With(mw.CachePrivate(1 * time.Minute)).Get("/dashboard/stats", handler)
func CachePrivate(maxAge time.Duration) func(http.Handler) http.Handler {
	v := fmt.Sprintf("private, max-age=%d", int(maxAge.Seconds()))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", v)
			next.ServeHTTP(w, r)
		})
	}
}
