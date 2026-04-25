package FUN

import (
	"errors"
	"fmt"
)

// ── Internal library errors ──────────────────────────────────────────────────

var (
	ErrSizeLimitExceeded = errors.New("size limit exceeded")
	ErrEncodingFailed    = errors.New("encoding failed")
	ErrInterceptorFailed = errors.New("interceptor error")
)

type SizeLimitError struct {
	Size int
	Max  int
}

func (e *SizeLimitError) Error() string {
	return fmt.Sprintf("response size (%d bytes) exceeds limit (%d bytes)", e.Size, e.Max)
}

type EncodingError struct {
	Inner error
}

func (e *EncodingError) Error() string {
	return fmt.Sprintf("encoding failed: %v", e.Inner)
}

type InterceptorLimitError struct {
	Current int
	Max     int
}

func (e *InterceptorLimitError) Error() string {
	return fmt.Sprintf("maximum number of interceptors reached: %d/%d", e.Current, e.Max)
}

type StatusCodeError struct {
	Code int
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("invalid HTTP status code: %d", e.Code)
}

// ── Request / param errors ───────────────────────────────────────────────────

// FieldError represents a single field-level validation failure.
// Used by both AppError (response) and Collector (request param extraction).
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e *FieldError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("(%s) %s: %v", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("(%s) %s", e.Field, e.Message)
}

// MissingParamError is returned when a required param is absent.
type MissingParamError struct {
	Key string
	Src string // "path" | "query" | "header" | "form"
}

func (e *MissingParamError) Error() string {
	return fmt.Sprintf("%s param %q is required", e.Src, e.Key)
}

// ParseError is returned when a param is present but cannot be coerced.
type ParseError struct {
	Key  string
	Src  string
	Got  string
	Want string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s param %q: expected %s, got %q", e.Src, e.Key, e.Want, e.Got)
}

// BodyError wraps a JSON decode failure from the request body.
type BodyError struct {
	Inner error
}

func (e *BodyError) Error() string {
	return fmt.Sprintf("invalid request body: %v", e.Inner)
}

func (e *BodyError) Unwrap() error { return e.Inner }

func IsBodyError(err error) bool {
	var t *BodyError
	return errors.As(err, &t)
}

func IsMissingParam(err error) bool {
	var t *MissingParamError
	return errors.As(err, &t)
}

func IsParseError(err error) bool {
	var t *ParseError
	return errors.As(err, &t)
}
