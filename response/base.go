package response

import (
	"log"
	"net/http"
	"time"
)

func newBaseResponse(code int, msg ...string) *Response {
	if err := validateStatusCode(code); err != nil {
		log.Printf("WARNING: newBaseResponse called with invalid status code %d: %v. Response will be sent as 500.", code, err)
    code = 500
	}

	var message string
	if len(msg) > 0 {
		message = msg[0]
	} else {
		message = ""
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

func Base(cfg ...*Config) *Response {
	var conf *Config
	if len(cfg) > 0 && cfg[0] != nil {
		conf = cfg[0]
	} else {
		c := getConfig()
		conf = &c
	}

	return &Response{
		ContentType: conf.DefaultContentType,
		config:      *conf,
	}
}

func (r *Response) applyMessage(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: applyMessage called on nil Response")
		return nil
	}
	if len(msg) > 0 {
		r.Message = msg[0]
	} else {
		r.Message = ""
	}
	return r
}

// Standard HTTP response builders
func OK(msg ...string) *Response {
	return newBaseResponse(http.StatusOK, msg...)
}
func Created(msg ...string) *Response {
	return newBaseResponse(http.StatusCreated, msg...)
}
func Accepted(msg ...string) *Response {
	return newBaseResponse(http.StatusAccepted, msg...)
}
func NoContent(msg ...string) *Response {
	return newBaseResponse(http.StatusNoContent, msg...)
}
func BadRequest(msg ...string) *Response {
	return newBaseResponse(http.StatusBadRequest, msg...)
}
func Unauthorized(msg ...string) *Response {
	return newBaseResponse(http.StatusUnauthorized, msg...)
}
func PaymentRequired(msg ...string) *Response {
	return newBaseResponse(http.StatusPaymentRequired, msg...)
}
func Forbidden(msg ...string) *Response {
	return newBaseResponse(http.StatusForbidden, msg...)
}
func NotFound(msg ...string) *Response {
	return newBaseResponse(http.StatusNotFound, msg...)
}
func MethodNotAllowed(msg ...string) *Response {
	return newBaseResponse(http.StatusMethodNotAllowed, msg...)
}
func Conflict(msg ...string) *Response {
	return newBaseResponse(http.StatusConflict, msg...)
}
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
func BadGateway(msg ...string) *Response {
	return newBaseResponse(http.StatusBadGateway, msg...)
}
func ServiceUnavailable(msg ...string) *Response {
	return newBaseResponse(http.StatusServiceUnavailable, msg...)
}

func (r *Response) OK(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: OK called on nil Response")
		return nil
	}
	r.Code = http.StatusOK
	r.applyMessage(msg...)
	return r
}
func (r *Response) Created(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: Created called on nil Response")
		return nil
	}
	r.Code = http.StatusCreated
	r.applyMessage(msg...)
	return r
}
func (r *Response) Accepted(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: Accepted called on nil Response")
		return nil
	}
	r.Code = http.StatusAccepted
	r.applyMessage(msg...)
	return r
}
func (r *Response) NoContent(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: NoContent called on nil Response")
		return nil
	}
	r.Code = http.StatusNoContent
	r.applyMessage(msg...)
	return r
}
func (r *Response) BadRequest(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: BadRequest called on nil Response")
		return nil
	}
	r.Code = http.StatusBadRequest
	r.applyMessage(msg...)
	return r
}
func (r *Response) Unauthorized(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: Unauthorized called on nil Response")
		return nil
	}
	r.Code = http.StatusUnauthorized
	r.applyMessage(msg...)
	return r
}
func (r *Response) PaymentRequired(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: PaymentRequired called on nil Response")
		return nil
	}
	r.Code = http.StatusPaymentRequired
	r.applyMessage(msg...)
	return r
}
func (r *Response) Forbidden(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: Forbidden called on nil Response")
		return nil
	}
	r.Code = http.StatusForbidden
	r.applyMessage(msg...)
	return r
}
func (r *Response) NotFound(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: NotFound called on nil Response")
		return nil
	}
	r.Code = http.StatusNotFound
	r.applyMessage(msg...)
	return r
}
func (r *Response) MethodNotAllowed(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: MethodNotAllowed called on nil Response")
		return nil
	}
	r.Code = http.StatusMethodNotAllowed
	r.applyMessage(msg...)
	return r
}
func (r *Response) Conflict(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: Conflict called on nil Response")
		return nil
	}
	r.Code = http.StatusConflict
	r.applyMessage(msg...)
	return r
}
func (r *Response) UnprocessableEntity(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: UnprocessableEntity called on nil Response")
		return nil
	}
	r.Code = http.StatusUnprocessableEntity
	r.applyMessage(msg...)
	return r
}
func (r *Response) TooManyRequests(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: TooManyRequests called on nil Response")
		return nil
	}
	r.Code = http.StatusTooManyRequests
	r.applyMessage(msg...)
	return r
}
func (r *Response) InternalServerError(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: InternalServerError called on nil Response")
		return nil
	}
	r.Code = http.StatusInternalServerError
	r.applyMessage(msg...)
	return r
}
func (r *Response) NotImplemented(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: NotImplemented called on nil Response")
		return nil
	}
	r.Code = http.StatusNotImplemented
	r.applyMessage(msg...)
	return r
}
func (r *Response) BadGateway(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: BadGateway called on nil Response")
		return nil
	}
	r.Code = http.StatusBadGateway
	r.applyMessage(msg...)
	return r
}
func (r *Response) ServiceUnavailable(msg ...string) *Response {
	if r == nil {
		log.Println("WARNING: ServiceUnavailable called on nil Response")
		return nil
	}
	r.Code = http.StatusServiceUnavailable
	r.applyMessage(msg...)
	return r
}
