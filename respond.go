package fun

import "net/http"

// Respond sends a JSON response with the given data.
// Uses 200 OK by default, or the optional status code.
//
// Usage:
//
//	fun.Respond(w, user)
//	fun.Respond(w, user, http.StatusCreated)
func Respond(w http.ResponseWriter, data any, code ...int) {
	status := http.StatusOK
	if len(code) > 0 {
		status = code[0]
	}
	base(status).WithData(data).Send(w)
}
