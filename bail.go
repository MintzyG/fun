package fun

import "net/http"

// Bail sends an error response and returns true if err is non-nil.
// Designed for handlers that can't adopt the HandlerFunc signature.
//
// Usage:
//
//	if fun.Bail(w, err) { return }
func Bail(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	Error(err).Send(w)
	return true
}
