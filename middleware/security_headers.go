package middleware

import "net/http"

// SecurityHeaders sets a baseline of security-related response headers suitable for APIs.
// For HTML-serving applications consider also adding Content-Security-Policy and Permissions-Policy.
//
//	r.Use(mw.SecurityHeaders())
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			// Prevent MIME-type sniffing
			h.Set("X-Content-Type-Options", "nosniff")
			// Block iframe embedding
			h.Set("X-Frame-Options", "DENY")
			// Control referrer leakage across origins
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			// Disable the broken legacy XSS filter
			h.Set("X-XSS-Protection", "0")
			next.ServeHTTP(w, r)
		})
	}
}
