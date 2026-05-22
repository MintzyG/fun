package fun

import "net/http"

// Bail sends an error response and returns true if err != nil.
// Pass an optional *http.Request to include the request path as RFC 7807 instance.
//
//	if fun.Bail(w, err) { return }         // simple error response
//	if fun.Bail(w, err, r) { return }      // RFC 7807 with instance
func Bail(w http.ResponseWriter, err error, r ...*http.Request) bool {
	if err == nil {
		return false
	}
	if len(r) > 0 && r[0] != nil {
		Problem(err, r[0]).Send(w)
		return true
	}
	Error(err).Send(w)
	return true
}
