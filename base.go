package fun

import (
	"log"
	"net/http"
	"time"
)

func newBaseResponse(code int, msg ...string) *Response {
	if err := validateStatusCode(code); err != nil {
		log.Printf("WARNING: newBaseResponse called with invalid status code %d: %v. Defaulting to 500.", code, err)
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

func OK(msg ...string) *Response           { return newBaseResponse(http.StatusOK, msg...) }
func Created(msg ...string) *Response      { return newBaseResponse(http.StatusCreated, msg...) }
func Accepted(msg ...string) *Response     { return newBaseResponse(http.StatusAccepted, msg...) }
func NoContent(msg ...string) *Response    { return newBaseResponse(http.StatusNoContent, msg...) }
func BadRequest(msg ...string) *Response   { return newBaseResponse(http.StatusBadRequest, msg...) }
func Unauthorized(msg ...string) *Response { return newBaseResponse(http.StatusUnauthorized, msg...) }
func PaymentRequired(msg ...string) *Response {
	return newBaseResponse(http.StatusPaymentRequired, msg...)
}
func Forbidden(msg ...string) *Response { return newBaseResponse(http.StatusForbidden, msg...) }
func NotFound(msg ...string) *Response  { return newBaseResponse(http.StatusNotFound, msg...) }
func MethodNotAllowed(msg ...string) *Response {
	return newBaseResponse(http.StatusMethodNotAllowed, msg...)
}
func Conflict(msg ...string) *Response { return newBaseResponse(http.StatusConflict, msg...) }
func UnprocessableEntity(msg ...string) *Response {
	return newBaseResponse(http.StatusUnprocessableEntity, msg...)
}
func TooManyRequests(msg ...string) *Response {
	return newBaseResponse(http.StatusTooManyRequests, msg...)
}
func InternalServerError(msg ...string) *Response {
	return newBaseResponse(http.StatusInternalServerError, msg...)
}
func NotImplemented(msg ...string) *Response {
	return newBaseResponse(http.StatusNotImplemented, msg...)
}
func BadGateway(msg ...string) *Response { return newBaseResponse(http.StatusBadGateway, msg...) }
func ServiceUnavailable(msg ...string) *Response {
	return newBaseResponse(http.StatusServiceUnavailable, msg...)
}
