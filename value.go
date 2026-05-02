package fun

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Value holds a raw string extracted from a path param, query param, or header,
// and exposes typed conversion methods.
//
// All typed methods return (T, error). The OrXxx variants return a fallback
// on missing/invalid input — great for optional params.
type Value struct {
	key     string
	raw     string
	src     string // "path" | "query" | "header"
	missing bool
}

// Raw returns the raw string value and whether it was present.
func (v Value) Raw() (string, bool) { return v.raw, !v.missing }

// String returns the raw value as-is, or "" if missing.
func (v Value) String() string { return v.raw }

// StringOr returns the raw value, or fallback if missing.
func (v Value) StringOr(fallback string) string {
	if v.missing || v.raw == "" {
		return fallback
	}
	return v.raw
}

// StringOpt returns the string value and whether it is present and non-empty.
// Returns ("", false) if the value is missing or empty.
//
// This is useful for optional parameters where you need to distinguish between
// "not provided" and an actual value without treating absence as an error.
func (v Value) StringOpt() (string, bool) {
	if v.missing || v.raw == "" {
		return "", false
	}
	return v.raw, true
}

// StringPtr returns a pointer to the string value, or nil if missing or empty.
//
// This is useful for optional fields where nil represents absence.
func (v Value) StringPtr() *string {
	str, ok := v.StringOpt()
	if !ok {
		return nil
	}
	return &str
}

// StringRequired returns the string value or a MissingParamError if absent or empty.
//
//	tok, err := req.Query("token").StringRequired()
//	if fun.Bail(w, err) { return }
func (v Value) StringRequired() (string, error) {
	if v.missing || v.raw == "" {
		return "", &MissingParamError{Key: v.key, Src: v.src}
	}
	return v.raw, nil
}

// Required returns an error if the value is missing or empty.
func (v Value) Required() error {
	if v.missing || v.raw == "" {
		return &MissingParamError{Key: v.key, Src: v.src}
	}
	return nil
}

// Int parses the value as int.
func (v Value) Int() (int, error) {
	if v.missing {
		return 0, &MissingParamError{Key: v.key, Src: v.src}
	}
	n, err := strconv.Atoi(v.raw)
	if err != nil {
		return 0, &ParseError{Key: v.key, Src: v.src, Got: v.raw, Want: "int"}
	}
	return n, nil
}

// IntOr parses the value as int, returning fallback on missing or parse error.
func (v Value) IntOr(fallback int) int {
	n, err := v.Int()
	if err != nil {
		return fallback
	}
	return n
}

// Int64 parses the value as int64.
func (v Value) Int64() (int64, error) {
	if v.missing {
		return 0, &MissingParamError{Key: v.key, Src: v.src}
	}
	n, err := strconv.ParseInt(v.raw, 10, 64)
	if err != nil {
		return 0, &ParseError{Key: v.key, Src: v.src, Got: v.raw, Want: "int64"}
	}
	return n, nil
}

// Int64Or parses the value as int64, returning fallback on error.
func (v Value) Int64Or(fallback int64) int64 {
	n, err := v.Int64()
	if err != nil {
		return fallback
	}
	return n
}

// Float64 parses the value as float64.
func (v Value) Float64() (float64, error) {
	if v.missing {
		return 0, &MissingParamError{Key: v.key, Src: v.src}
	}
	f, err := strconv.ParseFloat(v.raw, 64)
	if err != nil {
		return 0, &ParseError{Key: v.key, Src: v.src, Got: v.raw, Want: "float64"}
	}
	return f, nil
}

// Float64Or parses the value as float64, returning fallback on error.
func (v Value) Float64Or(fallback float64) float64 {
	f, err := v.Float64()
	if err != nil {
		return fallback
	}
	return f
}

// Bool parses "true"/"false"/"1"/"0" etc.
func (v Value) Bool() (bool, error) {
	if v.missing {
		return false, &MissingParamError{Key: v.key, Src: v.src}
	}
	b, err := strconv.ParseBool(v.raw)
	if err != nil {
		return false, &ParseError{Key: v.key, Src: v.src, Got: v.raw, Want: "bool"}
	}
	return b, nil
}

// BoolOr parses the value as bool, returning fallback on error.
func (v Value) BoolOr(fallback bool) bool {
	b, err := v.Bool()
	if err != nil {
		return fallback
	}
	return b
}

// UUID parses the value as a UUID.
func (v Value) UUID() (uuid.UUID, error) {
	if v.missing {
		return uuid.Nil, &MissingParamError{Key: v.key, Src: v.src}
	}
	id, err := uuid.Parse(v.raw)
	if err != nil {
		return uuid.Nil, &ParseError{Key: v.key, Src: v.src, Got: v.raw, Want: "uuid"}
	}
	return id, nil
}

// UUIDOr parses the value as UUID, returning fallback on error.
func (v Value) UUIDOr(fallback uuid.UUID) uuid.UUID {
	id, err := v.UUID()
	if err != nil || id == uuid.Nil {
		return fallback
	}
	return id
}

// UUIDOpt parses the value as a UUID and reports whether it is present and valid.
//
// It returns (uuid, true) only if the value is non-empty, parses successfully,
// and is not uuid.Nil. Otherwise, it returns (uuid.Nil, false).
//
// This is useful for optional parameters where you need to distinguish between
// “not provided / invalid” and a valid UUID without treating it as an error.
func (v Value) UUIDOpt() (uuid.UUID, bool) {
	if v.missing || v.raw == "" {
		return uuid.Nil, false
	}

	id, err := uuid.Parse(v.raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}

	return id, true
}

// UUIDPtr parses the value as a UUID and returns a pointer to it.
//
// It returns a non-nil *uuid.UUID only if the value is present, valid,
// and not uuid.Nil. Otherwise, it returns nil.
//
// This is useful for optional fields where nil represents absence.
func (v Value) UUIDPtr() *uuid.UUID {
	id, ok := v.UUIDOpt()
	if !ok {
		return nil
	}
	return &id
}

// Time parses the value using the given layout (e.g. time.RFC3339).
func (v Value) Time(layout string) (time.Time, error) {
	if v.missing {
		return time.Time{}, &MissingParamError{Key: v.key, Src: v.src}
	}
	t, err := time.Parse(layout, v.raw)
	if err != nil {
		return time.Time{}, &ParseError{Key: v.key, Src: v.src, Got: v.raw, Want: "time(" + layout + ")"}
	}
	return t, nil
}

// TimeOr parses the value as time, returning fallback on error.
func (v Value) TimeOr(layout string, fallback time.Time) time.Time {
	t, err := v.Time(layout)
	if err != nil {
		return fallback
	}
	return t
}

// Enum validates the value against an allowed set (case-sensitive).
// Returns a ParseError if the value is not in the set.
//
//	status := req.Query("status").Enum("active", "inactive", "pending")
func (v Value) Enum(allowed ...string) (string, error) {
	if v.missing {
		return "", &MissingParamError{Key: v.key, Src: v.src}
	}
	for _, a := range allowed {
		if v.raw == a {
			return v.raw, nil
		}
	}
	return "", &ParseError{
		Key:  v.key,
		Src:  v.src,
		Got:  v.raw,
		Want: "one of [" + strings.Join(allowed, ", ") + "]",
	}
}

// EnumOr validates the value against an allowed set, returning fallback if invalid.
func (v Value) EnumOr(fallback string, allowed ...string) string {
	s, err := v.Enum(allowed...)
	if err != nil {
		return fallback
	}
	return s
}

// StripBearer removes the "Bearer " prefix from Authorization headers.
// Returns the raw value unchanged if the prefix is absent.
//
//	tok := req.Header("Authorization").StripBearer()
func (v Value) StripBearer() string {
	return strings.TrimPrefix(v.raw, "Bearer ")
}

// IsPresent reports whether the value is non-empty.
func (v Value) IsPresent() bool { return !v.missing && v.raw != "" }

// IsMissing reports whether the value was absent.
func (v Value) IsMissing() bool { return v.missing }

// FieldErr converts a parse/missing error into a FieldError compatible
// with your response.AppError field-level detail.
//
//	id, err := req.Path("id").UUID()
//	if err != nil {
//	    ferr := request.From(r).Path("id").FieldErr(err)
//	    response.Error(response.NewError("bad id").WithFields(ferr).Validation()).Send(w)
//	}
func (v Value) FieldErr(err error) *FieldError {
	if err == nil {
		return nil
	}
	return &FieldError{
		Field:   v.src + "." + v.key,
		Message: err.Error(),
		Value:   v.raw,
	}
}
