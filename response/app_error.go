package response

import "fmt"

// Sentinel error codes and their HTTP status mappings.
const (
	ErrInternal   = "INTERNAL_ERROR"   // 500
	ErrValidation = "VALIDATION_ERROR" // 400
	ErrAuth       = "UNAUTHORIZED"     // 401
	ErrForbidden  = "FORBIDDEN"        // 403
	ErrNotFound   = "NOT_FOUND"        // 404
	ErrConflict   = "CONFLICT"         // 409
)

// appErrorStatusMap maps sentinel codes to HTTP status codes.
var appErrorStatusMap = map[string]int{
	ErrInternal:   500,
	ErrValidation: 400,
	ErrAuth:       401,
	ErrForbidden:  403,
	ErrNotFound:   404,
	ErrConflict:   409,
}

// FieldError represents a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
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
	Code string `json:"code"`

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
	if e.Status != 0 {
		return e.Status
	}
	if s, ok := appErrorStatusMap[e.Code]; ok {
		return s
	}
	return 500
}

// toResponse converts the AppError into a *Response ready to be sent.
// Debug is stripped here if not in development mode.
func (e *AppError) toResponse() *Response {
	config := getConfig()

	ae := &AppError{
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
		Meta:    e.Meta,
		// Debug handled below
	}

	if config.IsDevelopment && e.Debug != nil {
		ae.Debug = e.Debug
	}

	r := newBaseResponse(e.httpStatus(), e.Message)
	r.AppError = ae
	return r
}
