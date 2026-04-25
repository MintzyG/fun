package fun

import (
	"log"
	"net/http"
	"time"
)

func base(code int, msg ...string) *Response {
	if err := validateStatusCode(code); err != nil {
		log.Printf("[fun] WARNING: base called with invalid status code %d: %v. Defaulting to 500.", code, err)
		code = 500
	}

	var message string
	if len(msg) > 0 {
		message = msg[0]
	}

	config := getConfig()
	return &Response{
		Code:        code,
		Message:     message,
		Timestamp:   time.Now(),
		ContentType: config.DefaultContentType,
		Module:      config.DefaultModule,
	}
}

func OK(msg ...string) *Response                 { return base(http.StatusOK, msg...) }
func Created(msg ...string) *Response            { return base(http.StatusCreated, msg...) }
func Accepted(msg ...string) *Response           { return base(http.StatusAccepted, msg...) }
func NoContent(msg ...string) *Response          { return base(http.StatusNoContent, msg...) }
func BadRequest(msg ...string) *Response         { return base(http.StatusBadRequest, msg...) }
func Unauthorized(msg ...string) *Response       { return base(http.StatusUnauthorized, msg...) }
func PaymentRequired(msg ...string) *Response    { return base(http.StatusPaymentRequired, msg...) }
func Forbidden(msg ...string) *Response          { return base(http.StatusForbidden, msg...) }
func NotFound(msg ...string) *Response           { return base(http.StatusNotFound, msg...) }
func MethodNotAllowed(msg ...string) *Response   { return base(http.StatusMethodNotAllowed, msg...) }
func Conflict(msg ...string) *Response           { return base(http.StatusConflict, msg...) }
func TooManyRequests(msg ...string) *Response    { return base(http.StatusTooManyRequests, msg...) }
func NotImplemented(msg ...string) *Response     { return base(http.StatusNotImplemented, msg...) }
func BadGateway(msg ...string) *Response         { return base(http.StatusBadGateway, msg...) }
func ServiceUnavailable(msg ...string) *Response { return base(http.StatusServiceUnavailable, msg...) }
func UnprocessableEntity(msg ...string) *Response {
	return base(http.StatusUnprocessableEntity, msg...)
}
func InternalServerError(msg ...string) *Response {
	return base(http.StatusInternalServerError, msg...)
}
