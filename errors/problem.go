// Package fe provides RFC 9457 Problem Details for HTTP APIs.
//
// The central type is [Problem], which is both a Go error and a self-serializing
// HTTP response. It maps 1:1 onto the RFC 9457 JSON object and can be sent
// directly to an [http.ResponseWriter] with the correct Content-Type and status.
//
// Wire format (application/problem+json):
//
//	{
//	  "type":     "https://example.com/problems/out-of-credit",
//	  "title":    "Insufficient credit.",
//	  "status":   403,
//	  "detail":   "Your balance is 30, but that costs 50.",
//	  "instance": "/account/12345/msgs/abc",
//	  "balance":  30
//	}
//
// Extensions are flat on the object per §3.2 — there is no "meta" wrapper.
package fe

import (
	"encoding/json"
	"fmt"
)

// ProblemType is a URI reference that identifies a problem type.
//
// Per RFC 9457 §3.1.1, consumers MUST use the type URI as the primary
// identifier. When the URI is an http/https locator, dereferencing it
// SHOULD yield human-readable documentation for the problem type.
//
// Use [AboutBlank] when no additional semantics beyond the HTTP status
// code are needed. Define your own domain-specific URIs for everything else.
type ProblemType = string

const (
	// AboutBlank indicates the problem has no additional semantics beyond
	// the HTTP status code. Per RFC 9457 §4.2.1, title SHOULD match the
	// standard HTTP status phrase when this type is used.
	AboutBlank ProblemType = "about:blank"
)

// Problem is an RFC 9457 problem details object.
//
// It is both a Go error and a self-contained HTTP response. The zero value
// is not useful; construct via [New], [From], or the package-level helpers.
//
// Extensions (§3.2) are flat key/value pairs merged into the JSON object at
// serialization time. They must not use reserved names (type, title, status,
// detail, instance).
type Problem struct {
	// type (§3.1.1): URI identifying the problem type. Defaults to AboutBlank.
	typ ProblemType

	// title (§3.1.3): Short, human-readable summary of the problem type.
	// SHOULD NOT change between occurrences of the same type.
	title string

	// status (§3.1.2): HTTP status code. Advisory; must match the actual response code.
	status int

	// detail (§3.1.4): Human-readable explanation specific to this occurrence.
	// Should help the client correct the problem, not expose internals.
	detail string

	// instance (§3.1.5): URI identifying this specific occurrence.
	// May be a relative path (e.g. the request URI) or an absolute trace URI.
	instance string

	// extensions (§3.2): Flat extra members merged into the JSON object.
	extensions map[string]any

	// cause is the underlying Go error; not serialized.
	cause error
}

// reserved names that extensions must not overwrite.
var reserved = map[string]struct{}{
	"type": {}, "title": {}, "status": {}, "detail": {}, "instance": {},
}

// Error implements the error interface.
func (p *Problem) Error() string {
	if p.detail != "" {
		return fmt.Sprintf("[%d] %s: %s", p.status, p.title, p.detail)
	}
	return fmt.Sprintf("[%d] %s", p.status, p.title)
}

// Unwrap returns the underlying cause, enabling errors.Is/As chains.
func (p *Problem) Unwrap() error { return p.cause }

// Status returns the HTTP status code.
func (p *Problem) Status() int { return p.status }

// Type returns the problem type URI.
func (p *Problem) Type() ProblemType { return p.typ }

// Title returns the problem title.
func (p *Problem) Title() string { return p.title }

// Detail returns the occurrence-specific detail.
func (p *Problem) Detail() string { return p.detail }

// Instance returns the instance URI.
func (p *Problem) Instance() string { return p.instance }

// Extension returns the value of a named extension member, or nil.
func (p *Problem) Extension(key string) any {
	if p.extensions == nil {
		return nil
	}
	return p.extensions[key]
}

// MarshalJSON serializes the Problem as a flat RFC 9457 JSON object.
// Extensions are merged at the top level alongside the standard members.
func (p *Problem) MarshalJSON() ([]byte, error) {
	obj := make(map[string]any, 5+len(p.extensions))

	// Standard members — omit zero values per RFC guidance.
	typ := p.typ
	if typ == "" {
		typ = AboutBlank
	}
	obj["type"] = typ

	if p.title != "" {
		obj["title"] = p.title
	}
	if p.status != 0 {
		obj["status"] = p.status
	}
	if p.detail != "" {
		obj["detail"] = p.detail
	}
	if p.instance != "" {
		obj["instance"] = p.instance
	}

	// Extensions — flat, never overwrite reserved names.
	for k, v := range p.extensions {
		if _, clash := reserved[k]; !clash {
			obj[k] = v
		}
	}

	return json.Marshal(obj)
}
