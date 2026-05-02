package fun

import "net/http"

// HandlerFunc is like http.HandlerFunc but returns a *Response.
// The returned response is automatically sent — no need to call .Send(w).
// Returning nil is a no-op.
//
// Usage:
//
//	r.Get("/users/{id}", fun.Handler(GetUser))
//
//	func GetUser(w http.ResponseWriter, r *http.Request) *fun.Response {
//	    id, err := req.Path("id").UUID()
//	    if err != nil {
//	        return fun.Error(err)
//	    }
//	    return fun.OK().WithData(user)
//	}
type HandlerFunc func(http.ResponseWriter, *http.Request) *Response

// Handler wraps a HandlerFunc into a standard http.HandlerFunc.
func Handler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if resp := h(w, r); resp != nil {
			resp.Send(w)
		}
	}
}
