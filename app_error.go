package fun

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

// Sentinel error codes and their HTTP status mappings.
const (
	CodeBadRequest          ErrorCode = "BAD_REQUEST"
	CodeValidation          ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	CodePaymentRequired     ErrorCode = "PAYMENT_REQUIRED"
	CodeForbidden           ErrorCode = "FORBIDDEN"
	CodeNotFound            ErrorCode = "NOT_FOUND"
	CodeMethodNotAllowed    ErrorCode = "METHOD_NOT_ALLOWED"
	CodeConflict            ErrorCode = "CONFLICT"
	CodeUnprocessableEntity ErrorCode = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests     ErrorCode = "TOO_MANY_REQUESTS"
	CodeInternal            ErrorCode = "INTERNAL_ERROR"
	CodeNotImplemented      ErrorCode = "NOT_IMPLEMENTED"
	CodeBadGateway          ErrorCode = "BAD_GATEWAY"
	CodeServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
)

var appErrorStatusMap = map[ErrorCode]int{
	CodeBadRequest:          http.StatusBadRequest,
	CodeValidation:          http.StatusBadRequest,
	CodeUnauthorized:        http.StatusUnauthorized,
	CodePaymentRequired:     http.StatusPaymentRequired,
	CodeForbidden:           http.StatusForbidden,
	CodeNotFound:            http.StatusNotFound,
	CodeMethodNotAllowed:    http.StatusMethodNotAllowed,
	CodeConflict:            http.StatusConflict,
	CodeUnprocessableEntity: http.StatusUnprocessableEntity,
	CodeTooManyRequests:     http.StatusTooManyRequests,
	CodeInternal:            http.StatusInternalServerError,
	CodeNotImplemented:      http.StatusNotImplemented,
	CodeBadGateway:          http.StatusBadGateway,
	CodeServiceUnavailable:  http.StatusServiceUnavailable,
}

// DebugInfo holds development-only diagnostic information.
type DebugInfo struct {
	RawError   string `json:"raw_error,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// AppError is a structured application error that carries semantic meaning
// and maps directly onto an HTTP response shape.
type AppError struct {
	Type   ErrorCode      `json:"type"`
	Title  string         `json:"title"`
	Detail string         `json:"detail,omitempty"`
	Fields []FieldError   `json:"fields,omitempty"`
	Meta   map[string]any `json:"meta,omitempty"`
	Debug  *DebugInfo     `json:"debug,omitempty"`
	Err    error          `json:"-"`
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Type, e.Title) }

func (e *AppError) Unwrap() error { return e.Err }

// Is reports whether any error in the chain matches the given ErrorCode.
func Is(err error, code ErrorCode) bool {
	if appErr, ok := errors.AsType[*AppError](err); ok {
		return appErr.Type == code
	}
	return false
}

func (e *AppError) httpStatus() int {
	if s, ok := appErrorStatusMap[e.Type]; ok {
		return s
	}
	return http.StatusInternalServerError
}

func (e *AppError) toResponse() *Response {
	r := base(e.httpStatus(), e.Title)
	r.Type = e.Type
	r.Detail = e.Detail
	r.Fields = e.Fields
	r.Meta = e.Meta
	if e.Debug != nil {
		r.Meta["debug"] = e.Debug
	}
	return r
}

// ---------------------------------------------------------------------------
// Builder
// ---------------------------------------------------------------------------

type AppErrorBuilder struct {
	message string
	detail  string
	fields  []FieldError
	meta    map[string]any
	err     error
}

// Err starts building an AppError with a plain message.
func Err(message string) *AppErrorBuilder {
	return &AppErrorBuilder{message: message}
}

// Errf starts building an AppError with a formatted message.
func Errf(format string, args ...any) *AppErrorBuilder {
	return &AppErrorBuilder{message: fmt.Sprintf(format, args...)}
}

// WithFields attaches field-level validation errors.
// Accepts FieldError, *FieldError, or any error that wraps a *FieldError.
func (b *AppErrorBuilder) WithFields(fields ...any) *AppErrorBuilder {
	for _, f := range fields {
		switch v := f.(type) {
		case FieldError:
			b.fields = append(b.fields, v)
		case *FieldError:
			if v != nil {
				b.fields = append(b.fields, *v)
			}
		case error:
			if fe, ok := errors.AsType[*FieldError](v); ok {
				b.fields = append(b.fields, *fe)
			}
		}
	}
	return b
}

// WithMeta attaches arbitrary key/value context.
func (b *AppErrorBuilder) WithMeta(key string, value any) *AppErrorBuilder {
	if b.meta == nil {
		b.meta = make(map[string]any)
	}
	b.meta[key] = value
	return b
}

// WithErr attaches the underlying cause for errors.Is/As chains.
func (b *AppErrorBuilder) WithErr(err error) *AppErrorBuilder {
	b.err = err
	return b
}

func (b *AppErrorBuilder) Detail(detail string) *AppErrorBuilder {
	b.detail = detail
	return b
}

func (b *AppErrorBuilder) build(code ErrorCode) *AppError {
	return &AppError{
		Type:   code,
		Title:  b.message,
		Detail: b.detail,
		Fields: b.fields,
		Meta:   b.meta,
		Err:    b.err,
	}
}

// Builder terminal methods

func (b *AppErrorBuilder) BadRequest() *AppError          { return b.build(CodeBadRequest) }
func (b *AppErrorBuilder) Validation() *AppError          { return b.build(CodeValidation) }
func (b *AppErrorBuilder) Unauthorized() *AppError        { return b.build(CodeUnauthorized) }
func (b *AppErrorBuilder) PaymentRequired() *AppError     { return b.build(CodePaymentRequired) }
func (b *AppErrorBuilder) Forbidden() *AppError           { return b.build(CodeForbidden) }
func (b *AppErrorBuilder) NotFound() *AppError            { return b.build(CodeNotFound) }
func (b *AppErrorBuilder) MethodNotAllowed() *AppError    { return b.build(CodeMethodNotAllowed) }
func (b *AppErrorBuilder) Conflict() *AppError            { return b.build(CodeConflict) }
func (b *AppErrorBuilder) UnprocessableEntity() *AppError { return b.build(CodeUnprocessableEntity) }
func (b *AppErrorBuilder) TooManyRequests() *AppError     { return b.build(CodeTooManyRequests) }
func (b *AppErrorBuilder) Internal() *AppError            { return b.build(CodeInternal) }
func (b *AppErrorBuilder) NotImplemented() *AppError      { return b.build(CodeNotImplemented) }
func (b *AppErrorBuilder) BadGateway() *AppError          { return b.build(CodeBadGateway) }
func (b *AppErrorBuilder) ServiceUnavailable() *AppError  { return b.build(CodeServiceUnavailable) }

// ---------------------------------------------------------------------------
// Shorthand helpers — skip the builder when you just have a message
// ---------------------------------------------------------------------------

func ErrBadRequest(msg string) *AppError          { return Err(msg).BadRequest() }
func ErrValidation(msg string) *AppError          { return Err(msg).Validation() }
func ErrUnauthorized(msg string) *AppError        { return Err(msg).Unauthorized() }
func ErrPaymentRequired(msg string) *AppError     { return Err(msg).PaymentRequired() }
func ErrForbidden(msg string) *AppError           { return Err(msg).Forbidden() }
func ErrNotFound(msg string) *AppError            { return Err(msg).NotFound() }
func ErrMethodNotAllowed(msg string) *AppError    { return Err(msg).MethodNotAllowed() }
func ErrConflict(msg string) *AppError            { return Err(msg).Conflict() }
func ErrUnprocessableEntity(msg string) *AppError { return Err(msg).UnprocessableEntity() }
func ErrTooManyRequests(msg string) *AppError     { return Err(msg).TooManyRequests() }
func ErrInternal(msg string) *AppError            { return Err(msg).Internal() }
func ErrNotImplemented(msg string) *AppError      { return Err(msg).NotImplemented() }
func ErrBadGateway(msg string) *AppError          { return Err(msg).BadGateway() }
func ErrServiceUnavailable(msg string) *AppError  { return Err(msg).ServiceUnavailable() }

// Error resolves err into a *Response ready to be sent.
//
//	id, err := req.Path("id").UUID()
//	if err != nil {
//	    fun.Error(err).Send(w)
//	    return
//	}
//
// Resolution order:
//  1. err is already an *AppError — use it directly.
//  2. err is a *ParseError or *MissingParamError — mapped to a validation AppError.
//  3. err is a *BodyError — mapped to a bad request AppError.
//  4. A mapper is registered — delegate to it.
//  5. No mapper — log a warning and wrap as ErrInternal.
func Error(err error) *Response {
	return resolveAppError(err).toResponse()
}
