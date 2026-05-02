package fun

import (
	"context"
	"net/http"
	"strings"
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

// From wraps an *http.Request. Panics on nil - it's always a bug.
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

// QueryAll returns all values for a repeated and/or comma-separated query parameter.
//
//	?tags=a&tags=b        → ["a", "b"]
//	?tags=a,b,c           → ["a", "b", "c"]
//	?tags=a,b&tags=c,d    → ["a", "b", "c", "d"]
func (r *Request) QueryAll(key string) []string {
	raw := r.raw.URL.Query()[key]
	var result []string
	for _, v := range raw {
		for _, part := range strings.Split(v, ",") {
			if s := strings.TrimSpace(part); s != "" {
				result = append(result, s)
			}
		}
	}
	return result
}

// Header returns a Value for a request header.
//
//	tok := req.Header("Authorization").StripBearer()
func (r *Request) Header(key string) Value {
	raw := r.raw.Header.Get(key)
	return Value{key: key, raw: raw, src: "header", missing: raw == ""}
}

// Cookie returns a Value for a request cookie.
//
//	tok := req.Cookie("refresh_token").String()
func (r *Request) Cookie(name string) Value {
	c, err := r.raw.Cookie(name)
	if err != nil {
		return Value{key: name, raw: "", src: "cookie", missing: true}
	}
	return Value{key: name, raw: c.Value, src: "cookie", missing: c.Value == ""}
}

// Body returns a BodyReader for decoding the request body.
func (r *Request) Body() *BodyReader {
	return &BodyReader{r: r.raw}
}

// Into decodes the JSON body into dst.
// Pass true to reject unknown fields (strict mode).
//
//	if err := req.Into(&input); err != nil { ... }
//	if err := req.Into(&input, true); err != nil { ... }
func (r *Request) Into(dst any, exact ...bool) error {
	if len(exact) > 0 && exact[0] {
		return r.Body().IntoExact(dst)
	}
	return r.Body().Into(dst)
}

// ClientIP returns the real client IP, checking in order:
// CF-Connecting-IP → X-Forwarded-For (leftmost) → RemoteAddr
func (r *Request) ClientIP() string {
	if ip := r.raw.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if xff := r.raw.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	// strip port from RemoteAddr
	addr := r.raw.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}
