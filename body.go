package fun

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// BodyReader provides fluent body access.
type BodyReader struct {
	r       *http.Request
	maxSize int64 // 0 = no limit
}

// Limit caps how many bytes are read from the body.
// Returns a new BodyReader — original is unchanged.
//
//	req.Body().Limit(1 << 20).Into(&payload) // 1 MB cap
func (b *BodyReader) Limit(maxBytes int64) *BodyReader {
	return &BodyReader{r: b.r, maxSize: maxBytes}
}

func (b *BodyReader) reader() io.Reader {
	if b.maxSize > 0 {
		return io.LimitReader(b.r.Body, b.maxSize)
	}
	return b.r.Body
}

// Into decodes the JSON body into dst, ignoring unknown fields.
//
//	var in CreateUserInput
//	if err := req.Body().Into(&in); err != nil { ... }
func (b *BodyReader) Into(dst any) error {
	dec := json.NewDecoder(b.reader())
	if err := dec.Decode(dst); err != nil {
		return &BodyError{Inner: err}
	}
	return nil
}

// IntoStrict decodes the JSON body into dst, returning an error on unknown fields.
func (b *BodyReader) IntoStrict(dst any) error {
	dec := json.NewDecoder(b.reader())
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &BodyError{Inner: err}
	}
	return nil
}

// Bytes reads the raw body bytes.
func (b *BodyReader) Bytes() ([]byte, error) {
	data, err := io.ReadAll(b.reader())
	if err != nil {
		return nil, fmt.Errorf("request: failed to read body: %w", err)
	}
	return data, nil
}

// Text reads the body as a plain string.
func (b *BodyReader) Text() (string, error) {
	data, err := b.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Form parses application/x-www-form-urlencoded or multipart/form-data.
// After calling Form, use req.Raw().FormValue / req.Raw().PostForm.
func (b *BodyReader) Form() error {
	if err := b.r.ParseForm(); err != nil {
		return fmt.Errorf("request: failed to parse form: %w", err)
	}
	return nil
}

// MultipartForm parses a multipart form with the given max memory.
// After calling MultipartForm, use req.Raw().MultipartForm.
func (b *BodyReader) MultipartForm(maxMemory int64) (*multipart.Form, error) {
	if err := b.r.ParseMultipartForm(maxMemory); err != nil {
		return nil, fmt.Errorf("request: failed to parse multipart form: %w", err)
	}
	return b.r.MultipartForm, nil
}

// FormValue returns a named form field value after Form() has been called.
func (b *BodyReader) FormValue(key string) Value {
	raw := b.r.FormValue(key)
	return Value{key: key, raw: raw, src: "form", missing: raw == ""}
}
