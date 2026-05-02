package fun

import "github.com/google/uuid"

// Collector accumulates field errors from multiple param extractions
// and lets you flush them all at once as a *AppError.
//
// Usage:
//
//	c := fun.Collect(req)
//	id   := c.Path("id").UUID()
//	page := c.Query("page").IntOr(1)
//	if c.HasErrors() {
//	    fun.Error(c.Fail()).Send(w)
//	    return
//	}
type Collector struct {
	req    *Request
	errors []*FieldError
}

// Collect wraps a Request in a Collector for batch param extraction.
func Collect(req *Request) *Collector {
	return &Collector{req: req}
}

type collectedValue struct {
	v   Value
	col *Collector
}

func (c *Collector) collect(v Value) *collectedValue {
	return &collectedValue{v: v, col: c}
}

func (cv *collectedValue) record(err error) {
	if err == nil {
		return
	}
	cv.col.errors = append(cv.col.errors, cv.v.FieldErr(err))
}

// Path returns a collectedValue for a path param.
func (c *Collector) Path(key string) *collectedValue {
	return c.collect(c.req.Path(key))
}

// Query returns a collectedValue for a query param.
func (c *Collector) Query(key string) *collectedValue {
	return c.collect(c.req.Query(key))
}

// Header returns a collectedValue for a header.
func (c *Collector) Header(key string) *collectedValue {
	return c.collect(c.req.Header(key))
}

// Cookie returns a collectedValue for a cookie.
func (c *Collector) Cookie(name string) *collectedValue {
	return c.collect(c.req.Cookie(name))
}

// HasErrors reports whether any errors have been collected.
func (c *Collector) HasErrors() bool { return len(c.errors) > 0 }

// Fail returns an *AppError with all collected field errors, ready for Error().
// Returns nil if no errors were collected.
func (c *Collector) Fail() *AppError {
	if len(c.errors) == 0 {
		return nil
	}
	fields := make([]any, len(c.errors))
	for i, e := range c.errors {
		fields[i] = e
	}
	return Err("invalid params").WithFields(fields...).Validation()
}

// ── collectedValue typed extractors ─────────────────────────────────────────

func (cv *collectedValue) String() string {
	if err := cv.v.Required(); err != nil {
		cv.record(err)
		return ""
	}
	return cv.v.String()
}

func (cv *collectedValue) StringOr(fallback string) string {
	return cv.v.StringOr(fallback)
}

func (cv *collectedValue) StringOpt() (string, bool) {
	return cv.v.StringOpt()
}

func (cv *collectedValue) StringPtr() *string {
	return cv.v.StringPtr()
}

func (cv *collectedValue) StringRequired() (string, error) {
	return cv.v.StringRequired()
}

func (cv *collectedValue) Int() int {
	n, err := cv.v.Int()
	cv.record(err)
	return n
}

func (cv *collectedValue) IntOr(fallback int) int {
	return cv.v.IntOr(fallback)
}

func (cv *collectedValue) Int64() int64 {
	n, err := cv.v.Int64()
	cv.record(err)
	return n
}

func (cv *collectedValue) Int64Or(fallback int64) int64 {
	return cv.v.Int64Or(fallback)
}

func (cv *collectedValue) Bool() bool {
	b, err := cv.v.Bool()
	cv.record(err)
	return b
}

func (cv *collectedValue) BoolOr(fallback bool) bool {
	return cv.v.BoolOr(fallback)
}

func (cv *collectedValue) UUID() uuid.UUID {
	id, err := cv.v.UUID()
	cv.record(err)
	return id
}

func (cv *collectedValue) UUIDOr(fallback uuid.UUID) uuid.UUID {
	return cv.v.UUIDOr(fallback)
}

func (cv *collectedValue) UUIDOpt() (uuid.UUID, bool) {
	return cv.v.UUIDOpt()
}

func (cv *collectedValue) UUIDPtr() *uuid.UUID {
	return cv.v.UUIDPtr()
}

func (cv *collectedValue) Enum(allowed ...string) string {
	s, err := cv.v.Enum(allowed...)
	cv.record(err)
	return s
}

func (cv *collectedValue) EnumOr(fallback string, allowed ...string) string {
	return cv.v.EnumOr(fallback, allowed...)
}

func (cv *collectedValue) StripBearer() string {
	if err := cv.v.Required(); err != nil {
		cv.record(err)
		return ""
	}
	return cv.v.StripBearer()
}
