package response

import (
	"errors"
	"fmt"
)

type ErrorCode string

// Sentinel error codes and their HTTP status mappings.
const (
	ErrBadRequest   ErrorCode = "BAD_REQUEST"      // HTTP 400 | gRPC INVALID_ARGUMENT (3)
	ErrValidation   ErrorCode = "VALIDATION_ERROR" // HTTP 400 | gRPC INVALID_ARGUMENT (3)
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"     // HTTP 401 | gRPC UNAUTHENTICATED 16
	ErrForbidden    ErrorCode = "FORBIDDEN"        // HTTP 403 | gRPC PERMISSION_DENIED 7
	ErrNotFound     ErrorCode = "NOT_FOUND"        // HTTP 404 | gRPC NOT_FOUND 5
	ErrConflict     ErrorCode = "CONFLICT"         // HTTP 409 | gRPC ALREADY_EXISTS 6
	ErrInternal     ErrorCode = "INTERNAL_ERROR"   // HTTP 500 | gRPC INTERNAL 13
)

// appErrorStatusMap maps sentinel codes to HTTP status codes.
var appErrorStatusMap = map[ErrorCode]int{
	ErrValidation:   400,
	ErrUnauthorized: 401,
	ErrForbidden:    403,
	ErrNotFound:     404,
	ErrConflict:     409,
	ErrInternal:     500,
}

// FieldError represents a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e FieldError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("(%s) %s: %v", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("(%s) %s", e.Field, e.Message)
}

// DebugInfo holds development-only diagnostic information.
// It is stripped from responses when Config.IsDevelopment is false.
type DebugInfo struct {
	RawError   string `json:"raw_error,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// AppError is a structured application error that carries semantic meaning
// and maps directly onto an HTTP response shape.
//
// It implements the error interface so it can flow through normal Go error
// handling and be detected with errors.As.
type AppError struct {
	// Code is a machine-readable sentinel (e.g. ErrNotFound).
	Code ErrorCode `json:"code"`

	// Message is the human-readable description sent to the client.
	Message string `json:"message"`

	// Fields holds field-level detail, typically for validation errors.
	Fields []FieldError `json:"fields,omitempty"`

	// Meta holds arbitrary key/value context (request IDs, resource names, …).
	Meta map[string]any `json:"meta,omitempty"`

	// Debug is only included in responses when Config.IsDevelopment is true.
	Debug *DebugInfo `json:"debug,omitempty"`

	// Err is the underlying cause; used by errors.Is / errors.As chains.
	Err error `json:"-"`

	// Status is the HTTP status code. Derived automatically from Code via
	// appErrorStatusMap; falls back to 500 for unknown codes.
	Status int `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// httpStatus returns the HTTP status for this error, defaulting to 500.
func (e *AppError) httpStatus() int {
	if s, ok := appErrorStatusMap[e.Code]; ok {
		return s
	}
	return 500
}

// toResponse converts the AppError into a *Response ready to be sent.
func (e *AppError) toResponse() *Response {
	r := newBaseResponse(e.httpStatus(), e.Message)
	r.AppError = e
	return r
}

// AppErrorBuilder is an intermediate type that holds the message
// before the error code is chosen.
type AppErrorBuilder struct {
	message string
	fields  []FieldError
	meta    map[string]any
	err     error
}

// NewError starts building an AppError with a plain message.
func NewError(message string) *AppErrorBuilder {
	return &AppErrorBuilder{message: message}
}

// NewErrorf starts building an AppError with a formatted message.
func NewErrorf(format string, args ...any) *AppErrorBuilder {
	return &AppErrorBuilder{message: fmt.Sprintf(format, args...)}
}

func (b *AppErrorBuilder) WithErrors(errs ...error) *AppErrorBuilder {
	for _, err := range errs {
		b.WithFields(err)
	}
	return b
}

// WithFields attaches field-level validation errors.
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
			var fe *FieldError
			if errors.As(v, &fe) {
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

func (b *AppErrorBuilder) build(code ErrorCode) *AppError {
	return &AppError{
		Code:    code,
		Message: b.message,
		Fields:  b.fields,
		Meta:    b.meta,
		Err:     b.err,
	}
}

// BadRequest HTTP 400 | gRPC INVALID_ARGUMENT (3)
// Use for malformed requests. Use Validation() instead when you have field-level detail.
func (b *AppErrorBuilder) BadRequest() *AppError {
	return b.build(ErrBadRequest)
}

// Validation HTTP 400 | gRPC INVALID_ARGUMENT (3)
func (b *AppErrorBuilder) Validation() *AppError {
	return b.build(ErrValidation)
}

// Unauthorized HTTP 401 | gRPC UNAUTHENTICATED (16)
func (b *AppErrorBuilder) Unauthorized() *AppError {
	return b.build(ErrUnauthorized)
}

// Forbidden HTTP 403 | gRPC PERMISSION_DENIED (7)
func (b *AppErrorBuilder) Forbidden() *AppError {
	return b.build(ErrForbidden)
}

// NotFound HTTP 404 | gRPC NOT_FOUND (5)
func (b *AppErrorBuilder) NotFound() *AppError {
	return b.build(ErrNotFound)
}

// Conflict HTTP 409 | gRPC ALREADY_EXISTS (6)
func (b *AppErrorBuilder) Conflict() *AppError {
	return b.build(ErrConflict)
}

// Internal HTTP 500 | gRPC INTERNAL (13)
func (b *AppErrorBuilder) Internal() *AppError {
	return b.build(ErrInternal)
}
