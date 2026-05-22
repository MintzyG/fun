package fun

import "net/http"

// Problem resolves err into a RFC 7807-compliant *Response with the request
// path set as the instance field.
//
// Use Problem over Error when you want full RFC 7807 output including instance.
//
//	if err != nil {
//	    fun.Problem(err, r).Send(w)
//	    return
//	}
func Problem(err error, r *http.Request) *Response {
	resp := resolveAppError(err).toResponse()
	if r != nil {
		resp.Instance = r.URL.Path
	}
	return resp
}
