package fe

import "net/http"

// Builder constructs a [Problem] fluently.
//
// Acquire one via [New] (with a type URI) or [From] (wrapping an existing error).
// Call chaining methods to set fields, then terminate with a status-code method
// or [Builder.Build].
//
//	prob := errors.New("https://example.com/probs/out-of-credit").
//	    Title("Insufficient credit.").
//	    Detail("Your balance is 30, but that costs 50.").
//	    With("balance", 30).
//	    With("accounts", []string{"/account/12345"}).
//	    Forbidden()
//
//	prob.Send(w)
type Builder struct {
	p Problem
}

// New returns a Builder with the given problem type URI.
//
// Use [AboutBlank] when no additional semantics beyond the HTTP status
// code are needed. Prefer a stable, domain-specific URI otherwise.
//
//	errors.New("https://api.example.com/problems/insufficient-credit")
//	errors.New(errors.AboutBlank)
func New(typ ProblemType) *Builder {
	return &Builder{p: Problem{typ: typ}}
}

// From returns a Builder that wraps cause as the underlying Go error.
// The problem type defaults to [AboutBlank]; override with [Builder.Type].
func From(cause error) *Builder {
	return &Builder{p: Problem{typ: AboutBlank, cause: cause}}
}

// Type sets the problem type URI.
func (b *Builder) Type(typ ProblemType) *Builder {
	b.p.typ = typ
	return b
}

// Title sets the problem title (§3.1.3).
//
// Should be a stable, human-readable summary of the problem type —
// not occurrence-specific. Use [Builder.Detail] for per-occurrence text.
func (b *Builder) Title(title string) *Builder {
	b.p.title = title
	return b
}

// Detail sets the occurrence-specific detail (§3.1.4).
//
// Should help the client correct the problem, not expose server internals.
func (b *Builder) Detail(detail string) *Builder {
	b.p.detail = detail
	return b
}

// Instance sets the instance URI (§3.1.5).
//
// Identifies this specific occurrence. Typically set to the request URI
// (relative) or an absolute trace/correlation URI.
//
//	errors.New(...).Instance(r.URL.Path)
//	errors.New(...).Instance("/traces/01HXQ7ZRYD...")
func (b *Builder) Instance(uri string) *Builder {
	b.p.instance = uri
	return b
}

// Cause sets the underlying Go error without affecting serialization.
// Enables errors.Is/As chains through the Problem.
func (b *Builder) Cause(err error) *Builder {
	b.p.cause = err
	return b
}

// With adds a flat extension member (§3.2).
//
// Extensions are merged into the JSON object at the top level alongside
// the standard members. Attempting to set a reserved name (type, title,
// status, detail, instance) is silently ignored.
//
// Names SHOULD start with a letter, SHOULD be ≥3 characters, and SHOULD
// comprise only letters, digits, and underscores (RFC 9457 §4).
//
//	.With("balance", 30)
//	.With("accounts", []string{"/account/12345", "/account/67890"})
func (b *Builder) With(key string, value any) *Builder {
	if _, clash := reserved[key]; clash {
		return b
	}
	if b.p.extensions == nil {
		b.p.extensions = make(map[string]any)
	}
	b.p.extensions[key] = value
	return b
}

// WithExtensions merges multiple extension members at once.
// Reserved names are silently skipped.
func (b *Builder) WithExtensions(ext map[string]any) *Builder {
	for k, v := range ext {
		b.With(k, v)
	}
	return b
}

// Build materializes the Problem with an explicit status code.
// Use when constructing via New or From directly.
//
//	errors.New("https://...").Title("...").Build(http.StatusTeapot)
func (b *Builder) Build(status int) *Problem {
	b.p.status = status
	return new(b.p)
}

// Err materializes the Problem using the preloaded status.
// Use with NewXxx helpers which pre-set the status via newBlank.
//
//	errors.NewNotFound().Type("https://...").Detail("...").Err()
func (b *Builder) Err() *Problem {
	return new(b.p)
}

// ── Status-code terminal methods ─────────────────────────────────────────────

func (b *Builder) Continue() *Problem         { return b.Build(http.StatusContinue) }
func (b *Builder) OK() *Problem               { return b.Build(http.StatusOK) }
func (b *Builder) Created() *Problem          { return b.Build(http.StatusCreated) }
func (b *Builder) Accepted() *Problem         { return b.Build(http.StatusAccepted) }
func (b *Builder) BadRequest() *Problem       { return b.Build(http.StatusBadRequest) }
func (b *Builder) Unauthorized() *Problem     { return b.Build(http.StatusUnauthorized) }
func (b *Builder) PaymentRequired() *Problem  { return b.Build(http.StatusPaymentRequired) }
func (b *Builder) Forbidden() *Problem        { return b.Build(http.StatusForbidden) }
func (b *Builder) NotFound() *Problem         { return b.Build(http.StatusNotFound) }
func (b *Builder) MethodNotAllowed() *Problem { return b.Build(http.StatusMethodNotAllowed) }
func (b *Builder) Conflict() *Problem         { return b.Build(http.StatusConflict) }
func (b *Builder) Gone() *Problem             { return b.Build(http.StatusGone) }
func (b *Builder) UnprocessableEntity() *Problem {
	return b.Build(http.StatusUnprocessableEntity)
}
func (b *Builder) TooManyRequests() *Problem { return b.Build(http.StatusTooManyRequests) }
func (b *Builder) InternalServerError() *Problem {
	return b.Build(http.StatusInternalServerError)
}
func (b *Builder) NotImplemented() *Problem     { return b.Build(http.StatusNotImplemented) }
func (b *Builder) BadGateway() *Problem         { return b.Build(http.StatusBadGateway) }
func (b *Builder) ServiceUnavailable() *Problem { return b.Build(http.StatusServiceUnavailable) }
func (b *Builder) GatewayTimeout() *Problem     { return b.Build(http.StatusGatewayTimeout) }
