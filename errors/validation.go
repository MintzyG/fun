package fe

import (
	"net/http"
	"strings"
)

// ValidationError describes a single field-level validation failure.
//
// RFC 9457 §3.2 and the spec's own validation example use an "errors"
// extension with "detail" and "pointer" (JSON Pointer, RFC 6901) members.
//
//	{
//	  "type":   "https://example.net/validation-error",
//	  "title":  "Your request is not valid.",
//	  "status": 422,
//	  "errors": [
//	    { "detail": "must be a positive integer", "pointer": "#/age" },
//	    { "detail": "must be 'green', 'red' or 'blue'", "pointer": "#/profile/color" }
//	  ]
//	}
type ValidationError struct {
	// Pointer is a JSON Pointer (RFC 6901) to the field that caused the error.
	// Examples: "#/age", "#/profile/color", "#/address/street"
	// For request param errors the convention is: "#/query/page", "#/path/id", "#/header/Authorization"
	Pointer string `json:"pointer"`

	// Detail describes this specific violation in actionable terms.
	Detail string `json:"detail"`
}

// ValidationBuilder constructs a validation Problem with the RFC 9457
// "errors" extension.
//
// Acquire via [Validation].
type ValidationBuilder struct {
	b      *Builder
	errors []ValidationError
}

// Validation returns a ValidationBuilder for a 422 Unprocessable Content response.
//
// Use [ValidationBuilder.Add] or [ValidationBuilder.Field] to accumulate errors,
// then call [ValidationBuilder.Build] or [ValidationBuilder.Send].
//
//	errors.Validation("https://api.example.com/problems/validation-error").
//	    Title("Your request is not valid.").
//	    Field("#/age", "must be a positive integer").
//	    Field("#/profile/color", "must be 'green', 'red' or 'blue'").
//	    Send(w)
func Validation(typ ProblemType) *ValidationBuilder {
	return &ValidationBuilder{
		b: New(typ),
	}
}

// Title sets the problem title. Defaults to "Your request is not valid."
// if not called.
func (vb *ValidationBuilder) Title(title string) *ValidationBuilder {
	vb.b.Title(title)
	return vb
}

// Detail sets the occurrence-level detail on the Problem.
func (vb *ValidationBuilder) Detail(detail string) *ValidationBuilder {
	vb.b.Detail(detail)
	return vb
}

// Instance sets the instance URI on the Problem.
func (vb *ValidationBuilder) Instance(uri string) *ValidationBuilder {
	vb.b.Instance(uri)
	return vb
}

// Add appends a [ValidationError] with an explicit pointer and detail.
//
// The pointer should be a valid JSON Pointer (RFC 6901).
//
//	.Add("#/age", "must be a positive integer")
//	.Add("#/profile/color", "must be 'green', 'red' or 'blue'")
func (vb *ValidationBuilder) Add(pointer, detail string) *ValidationBuilder {
	vb.errors = append(vb.errors, ValidationError{
		Pointer: pointer,
		Detail:  detail,
	})
	return vb
}

// Field is a convenience wrapper around [ValidationBuilder.Add] that builds
// the JSON Pointer from a source and field name.
//
//	.Field("body",   "email",  "must be a valid email address")
//	.Field("query",  "page",   "must be a positive integer")
//	.Field("path",   "id",     "must be a valid UUID")
//	.Field("header", "Authorization", "is required")
func (vb *ValidationBuilder) Field(src, name, detail string) *ValidationBuilder {
	pointer := "#/" + strings.ToLower(src) + "/" + name
	return vb.Add(pointer, detail)
}

// HasErrors reports whether any validation errors have been accumulated.
func (vb *ValidationBuilder) HasErrors() bool {
	return len(vb.errors) > 0
}

// Build materializes the [Problem]. The "errors" extension is set only if
// there are accumulated validation errors.
//
// Status defaults to 422 Unprocessable Content.
func (vb *ValidationBuilder) Build() *Problem {
	title := vb.b.p.title
	if title == "" {
		vb.b.Title("Your request is not valid.")
	}
	if len(vb.errors) > 0 {
		vb.b.With("errors", vb.errors)
	}
	return vb.b.Build(http.StatusUnprocessableEntity)
}
