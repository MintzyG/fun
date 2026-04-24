package middlewares

import (
	"net"
	"net/http"
	"strings"
)

// RealIP rewrites r.RemoteAddr with the real client IP, in order:
//  1. CF-Connecting-IP — set by Cloudflare, cannot be spoofed by clients
//  2. X-Forwarded-For — leftmost (client) IP, used for direct/non-Cloudflare traffic
//  3. RemoteAddr — unchanged fallback for direct connections with no proxy headers
//
// Place this as early as possible in the middleware chain so that downstream
// middleware (e.g. RateLimit) see the correct IP.
//
//	r.Use(mw.RealIP())
func RealIP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := r.Header.Get("CF-Connecting-IP"); validIP(ip) {
				r.RemoteAddr = ip
			} else if ip := leftmostXFF(r.Header.Get("X-Forwarded-For")); validIP(ip) {
				r.RemoteAddr = ip
			}
			// else: leave RemoteAddr as-is
			next.ServeHTTP(w, r)
		})
	}
}

// leftmostXFF returns the first (client) IP from a comma-separated X-Forwarded-For value.
// X-Forwarded-For: <client>, <proxy1>, <proxy2>
func leftmostXFF(xff string) string {
	if xff == "" {
		return ""
	}
	first, _, _ := strings.Cut(xff, ",")
	return strings.TrimSpace(first)
}

// validIP reports whether s is a valid, non-empty IP address.
func validIP(s string) bool {
	return s != "" && net.ParseIP(s) != nil
}
