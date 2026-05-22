package fe

import "net/http"

// ── Generic helpers — about:blank, returns *Problem directly ─────────────────
//
// Use when no additional semantics beyond the HTTP status are needed.
// Title defaults to the standard HTTP status phrase (RFC 9457 §4.2.1).
//
//	return errors.NotFound("user not found")

func blank(status int, title string) *Problem {
	if title == "" {
		title = http.StatusText(status)
	}
	return New(AboutBlank).Title(title).Build(status)
}

func BadRequest(title ...string) *Problem   { return blank(http.StatusBadRequest, first(title)) }
func Unauthorized(title ...string) *Problem { return blank(http.StatusUnauthorized, first(title)) }
func PaymentRequired(title ...string) *Problem {
	return blank(http.StatusPaymentRequired, first(title))
}
func Forbidden(title ...string) *Problem { return blank(http.StatusForbidden, first(title)) }
func NotFound(title ...string) *Problem  { return blank(http.StatusNotFound, first(title)) }
func MethodNotAllowed(title ...string) *Problem {
	return blank(http.StatusMethodNotAllowed, first(title))
}
func Conflict(title ...string) *Problem { return blank(http.StatusConflict, first(title)) }
func Gone(title ...string) *Problem     { return blank(http.StatusGone, first(title)) }
func UnprocessableEntity(title ...string) *Problem {
	return blank(http.StatusUnprocessableEntity, first(title))
}
func TooManyRequests(title ...string) *Problem {
	return blank(http.StatusTooManyRequests, first(title))
}
func Internal(title ...string) *Problem       { return blank(http.StatusInternalServerError, first(title)) }
func NotImplemented(title ...string) *Problem { return blank(http.StatusNotImplemented, first(title)) }
func BadGateway(title ...string) *Problem     { return blank(http.StatusBadGateway, first(title)) }
func ServiceUnavailable(title ...string) *Problem {
	return blank(http.StatusServiceUnavailable, first(title))
}
func GatewayTimeout(title ...string) *Problem { return blank(http.StatusGatewayTimeout, first(title)) }

// ── Builder helpers — pre-loaded status, returns *Builder for chaining ────────
//
// Use when you need a domain-specific type URI, extensions, or detail.
//
//	return errors.NewNotFound().
//	    Type("https://docs.example.com/problems/missing-subject").
//	    Detail("subject was not found in request context").
//	    Build()

func newBlank(status int) *Builder {
	b := New(AboutBlank).Title(http.StatusText(status))
	b.p.status = status
	return b
}

func NewBadRequest() *Builder          { return newBlank(http.StatusBadRequest) }
func NewUnauthorized() *Builder        { return newBlank(http.StatusUnauthorized) }
func NewPaymentRequired() *Builder     { return newBlank(http.StatusPaymentRequired) }
func NewForbidden() *Builder           { return newBlank(http.StatusForbidden) }
func NewNotFound() *Builder            { return newBlank(http.StatusNotFound) }
func NewMethodNotAllowed() *Builder    { return newBlank(http.StatusMethodNotAllowed) }
func NewConflict() *Builder            { return newBlank(http.StatusConflict) }
func NewGone() *Builder                { return newBlank(http.StatusGone) }
func NewUnprocessableEntity() *Builder { return newBlank(http.StatusUnprocessableEntity) }
func NewTooManyRequests() *Builder     { return newBlank(http.StatusTooManyRequests) }
func NewInternal() *Builder            { return newBlank(http.StatusInternalServerError) }
func NewNotImplemented() *Builder      { return newBlank(http.StatusNotImplemented) }
func NewBadGateway() *Builder          { return newBlank(http.StatusBadGateway) }
func NewServiceUnavailable() *Builder  { return newBlank(http.StatusServiceUnavailable) }
func NewGatewayTimeout() *Builder      { return newBlank(http.StatusGatewayTimeout) }

func first(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}
