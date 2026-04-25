// Package bind provides a decode+validate step for request bodies.
// It depends on go-playground/validator and is intentionally separate
// from the core request package so validator is an opt-in dependency.
//
// Setup (once at startup):
//
//	bind.SetValidator(validator.New())
//
// Usage:
//
//	var input CreateEventInput
//	if err := bind.Body(req).Bind(&input); err != nil {
//	    response.Error(err).Send(w)
//	    return
//	}
package bind

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/MintzyG/FastUtilitiesNet"
	"github.com/go-playground/validator/v10"
)

var (
	globalValidator *validator.Validate
	validatorMu     sync.RWMutex
)

// SetValidator registers the global validator instance.
// Call once at application startup.
//
//	v := validator.New()
//	v.RegisterTagNameFunc(func(f reflect.StructField) string {
//	    return strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
//	})
//	bind.SetValidator(v)
func SetValidator(v *validator.Validate) {
	validatorMu.Lock()
	defer validatorMu.Unlock()
	globalValidator = v
}

// Binder wraps a BodyReader and adds Bind.
type Binder struct {
	body *fun.BodyReader
}

// Body wraps a BodyReader for decode+validate.
func Body(req *fun.Request) *Binder {
	return &Binder{body: req.Body()}
}

// Limit caps how many bytes are read. Mirrors BodyReader.Limit.
func (b *Binder) Limit(maxBytes int64) *Binder {
	return &Binder{body: b.body.Limit(maxBytes)}
}

// Bind decodes the JSON body into dst and validates it.
// Lenient by default (unknown fields ignored). Pass true to enable strict mode.
// Returns nil on success, or a *AppError with field-level detail ready for
// response.Error().
//
//	if err := bind.Body(req).Bind(&input); err != nil {
//	    response.Error(err).Send(w)
//	    return
//	}
//
//	// strict mode — unknown fields are rejected
//	if err := bind.Body(req).Bind(&input, true); err != nil { ... }
func (b *Binder) Bind(dst any, strict ...bool) *fun.AppError {
	var decodeErr error
	if len(strict) > 0 && strict[0] {
		decodeErr = b.body.IntoStrict(dst)
	} else {
		decodeErr = b.body.Into(dst)
	}
	if decodeErr != nil {
		return fun.NewError(decodeErr.Error()).BadRequest()
	}

	validatorMu.RLock()
	v := globalValidator
	validatorMu.RUnlock()

	if v == nil {
		panic("bind: no validator registered — call bind.SetValidator at startup")
	}

	if err := v.Struct(dst); err != nil {
		fields := validationErrsToFields(err)
		return fun.NewError("invalid body").WithFields(fields...).Validation()
	}

	return nil
}

// validationErrsToFields converts validator.ValidationErrors into
// []any of *resp.FieldError, with password masking and kind-aware messages.
func validationErrsToFields(err error) []any {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []any{&fun.FieldError{Field: "body", Message: err.Error()}}
	}

	out := make([]any, 0, len(ve))
	for _, fe := range ve {
		if fe.Tag() == "passwd" {
			out = append(out, passwdFieldErrors(fe)...)
			continue
		}

		var value any
		if !isPasswordField(fe.Field()) {
			value = fe.Value()
		}

		out = append(out, &fun.FieldError{
			Field:   fe.Field(),
			Message: tagMessage(fe),
			Value:   value,
		})
	}
	return out
}

// isPasswordField reports whether a field name looks like a password field.
func isPasswordField(name string) bool {
	lower := strings.ToLower(name)
	for _, word := range []string{"password", "passwd", "pwd", "pass"} {
		if lower == word || strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// passwdFieldErrors expands a failed passwd tag into one FieldError per
// missing requirement.
func passwdFieldErrors(fe validator.FieldError) []any {
	password, ok := fe.Value().(string)
	if !ok {
		return []any{&fun.FieldError{Field: fe.Field(), Message: "must be a valid string"}}
	}

	var hasUpper, hasNumber, hasSymbol bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSymbol = true
		}
	}

	var out []any
	if !hasUpper {
		out = append(out, &fun.FieldError{Field: fe.Field(), Message: "must contain at least one uppercase letter"})
	}
	if !hasNumber {
		out = append(out, &fun.FieldError{Field: fe.Field(), Message: "must contain at least one number"})
	}
	if !hasSymbol {
		out = append(out, &fun.FieldError{Field: fe.Field(), Message: "must contain at least one symbol or punctuation"})
	}
	return out
}

// tagMessage produces a human-readable message for a validation failure.
// Kind-aware for min/max/gt/gte/lt/lte so strings say "characters" and
// numbers say the actual bound.
func tagMessage(fe validator.FieldError) string {
	tagMessagesMu.RLock()
	msg, ok := tagMessages[fe.Tag()]
	tagMessagesMu.RUnlock()
	if ok {
		return msg
	}

	param := fe.Param()
	isStr := fe.Kind() == reflect.String

	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "uuid4":
		return "must be a valid UUIDv4"
	case "uuid7":
		return "must be a valid UUIDv7"
	case "numeric":
		return "must be a numeric value"
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	case "len":
		if isStr {
			return fmt.Sprintf("must be exactly %s characters long", param)
		}
		return fmt.Sprintf("must have exactly %s elements", param)
	case "min":
		if isStr {
			return fmt.Sprintf("must be at least %s characters long", param)
		}
		return fmt.Sprintf("must be at least %s", param)
	case "max":
		if isStr {
			return fmt.Sprintf("must be at most %s characters long", param)
		}
		return fmt.Sprintf("must be at most %s", param)
	case "gt":
		if isStr {
			return fmt.Sprintf("must be longer than %s characters", param)
		}
		return fmt.Sprintf("must be greater than %s", param)
	case "gte":
		if isStr {
			return fmt.Sprintf("must be at least %s characters long", param)
		}
		return fmt.Sprintf("must be greater than or equal to %s", param)
	case "lt":
		if isStr {
			return fmt.Sprintf("must be shorter than %s characters", param)
		}
		return fmt.Sprintf("must be less than %s", param)
	case "lte":
		if isStr {
			return fmt.Sprintf("must be at most %s characters long", param)
		}
		return fmt.Sprintf("must be less than or equal to %s", param)
	case "oneof":
		opts := strings.ReplaceAll(param, " ", ", ")
		return fmt.Sprintf("must be one of: %s", opts)
	}

	return "failed validation: " + fe.Tag()
}

// tagMessages holds custom overrides registered via RegisterTagMessage.
var (
	tagMessages   = map[string]string{}
	tagMessagesMu sync.RWMutex
)

// RegisterTagMessage registers a custom human-readable message for a validator tag.
// Overrides the built-in kind-aware message for that tag entirely.
//
//	bind.RegisterTagMessage("required", "campo obrigatório")
func RegisterTagMessage(tag, message string) {
	tagMessagesMu.Lock()
	defer tagMessagesMu.Unlock()
	tagMessages[tag] = message
}
