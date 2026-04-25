package FUN

import (
	"context"
	"net/http"
)

// Request wraps an *http.Request and exposes typed, ergonomic accessors
// for path params, query params, headers, and the body.
//
// Usage:
//
//	req := fun.From(r)
//	id   := req.Path("id").UUID()
//	page := req.Query("page").IntOr(1)
//	tok  := req.Header("Authorization").StripBearer()
//	var body CreateUserInput
//	if err := req.Body().Into(&body); err != nil { ... }
type Request struct {
	raw        *http.Request
	pathParams PathParamFunc
}

// PathParamFunc abstracts router-specific path param extraction.
// Register one via SetPathParamFunc.
//
//	fun.SetPathParamFunc(chi.URLParam)
//	fun.SetPathParamFunc(func(r *http.Request, k string) string { return mux.Vars(r)[k] })
type PathParamFunc func(r *http.Request, key string) string

var globalPathParamFunc PathParamFunc

// SetPathParamFunc registers the router-specific path param extractor.
// Call once at application startup.
func SetPathParamFunc(fn PathParamFunc) {
	globalPathParamFunc = fn
}

// From wraps an *http.Request. Panics on nil — it's always a bug.
func From(r *http.Request) *Request {
	if r == nil {
		panic("fun.From: nil *http.Request")
	}
	return &Request{
		raw:        r,
		pathParams: globalPathParamFunc,
	}
}

// Raw returns the underlying *http.Request.
func (r *Request) Raw() *http.Request { return r.raw }

// Context returns the request context.
func (r *Request) Context() context.Context { return r.raw.Context() }

// Method returns the HTTP method (uppercased by net/http).
func (r *Request) Method() string { return r.raw.Method }

// Path returns a Value for a path parameter.
//
//	id := req.Path("id").UUID()
func (r *Request) Path(key string) Value {
	if r.pathParams == nil {
		return Value{key: key, raw: "", src: "path", missing: true}
	}
	raw := r.pathParams(r.raw, key)
	return Value{key: key, raw: raw, src: "path", missing: raw == ""}
}

// Query returns a Value for a query parameter.
//
//	page := req.Query("page").IntOr(1)
func (r *Request) Query(key string) Value {
	raw := r.raw.URL.Query().Get(key)
	return Value{key: key, raw: raw, src: "query", missing: raw == ""}
}

// QueryAll returns all values for a repeated query parameter.
//
//	tags := req.QueryAll("tag") // ?tag=a&tag=b → ["a", "b"]
func (r *Request) QueryAll(key string) []string {
	return r.raw.URL.Query()[key]
}

// Header returns a Value for a request header.
//
//	tok := req.Header("Authorization").StripBearer()
func (r *Request) Header(key string) Value {
	raw := r.raw.Header.Get(key)
	return Value{key: key, raw: raw, src: "header", missing: raw == ""}
}

// Body returns a BodyReader for decoding the request body.
func (r *Request) Body() *BodyReader {
	return &BodyReader{r: r.raw}
}
