package response

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Module         string          `json:"module,omitempty"`
	Message        string          `json:"message,omitempty"`
	Data           any             `json:"data,omitempty"`
	Trace          []string        `json:"trace,omitempty"`
	Timestamp      time.Time       `json:"timestamp,omitempty"`
	PaginationData *PaginationMeta `json:"pagination,omitempty"`
	Code           int             `json:"code,omitempty"`
	ErrorID        string          `json:"error_id,omitempty"`
	ContentType    string          `json:"-"`
	TracePrefix    string          `json:"-"`
	config         Config          `json:"-"`
}

// WithConfig sets a custom configuration for this specific response instance
// This overrides the global configuration for this response only
func (r *Response) WithConfig(config Config) *Response {
	if r == nil {
		log.Println("WARNING: WithConfig called on nil Response")
		return nil
	}
	// Validate and set defaults for invalid config values
	if config.MaxTraceSize <= 0 {
		config.MaxTraceSize = defaultConfig.MaxTraceSize
	}
	if config.ResponseSizeLimit <= 0 {
		config.ResponseSizeLimit = defaultConfig.ResponseSizeLimit
	}
	if config.MaxInterceptorAmount <= 0 {
		config.MaxInterceptorAmount = defaultConfig.MaxInterceptorAmount
	}
	if config.DefaultContentType == "" {
		config.DefaultContentType = defaultConfig.DefaultContentType
	}

	r.config = config

	// Update ContentType if it wasn't explicitly set
	if r.ContentType == "" || r.ContentType == getConfig().DefaultContentType {
		r.ContentType = config.DefaultContentType
	}

	return r
}

func (r *Response) WithContentType(ctype string) *Response {
	if r == nil {
		log.Println("WARNING: WithContentType called on nil Response")
		return nil
	}
	r.ContentType = ctype
	return r
}

func (r *Response) WithModule(module string) *Response {
	if r == nil {
		log.Println("WARNING: WithModule called on nil Response")
		return nil
	}
	r.Module = module
	return r
}

func (r *Response) WithMsg(message string) *Response {
	if r == nil {
		log.Println("WARNING: WithMsg called on nil Response")
		return nil
	}
	r.Message = message
	return r
}

func (r *Response) WithData(data any) *Response {
	if r == nil {
		log.Println("WARNING: WithData called on nil Response")
		return nil
	}
	r.Data = data
	return r
}

func (r *Response) WithTracePrefix(prefix string) *Response {
	if r == nil {
		log.Println("WARNING: WithTracePrefix called on nil Response")
		return nil
	}
	r.TracePrefix = prefix
	return r
}

func (r *Response) WithErrID(id string) *Response {
	if r == nil {
		log.Println("WARNING: WithErrID called on nil Response")
		return nil
	}
	r.ErrorID = id
	return r
}

// Does nothing unless using a custom response
func (r *Response) WithCode(code int) *Response {
	if r == nil {
		log.Println("WARNING: WithCode called on nil Response")
		return nil
	}
	if err := validateStatusCode(code); err != nil {
		log.Printf("WARNING: Invalid status code %d set. Response will not be sent.", code)
	}
	r.Code = code
	return r
}

// For when you don't have context (simple cases, tests, etc.)
func (r *Response) Send(w http.ResponseWriter) {
	if r == nil {
		log.Println("WARNING: Send called on nil Response")
		return
	}
	if w == nil {
		log.Println("WARNING: Send called with nil ResponseWriter")
		return
	}
	r.SendWithContext(context.Background(), w)
}

// For when you have context (web servers, etc.)
func (r *Response) SendWithContext(ctx context.Context, w http.ResponseWriter) {
	if r == nil {
		log.Println("WARNING: SendWithContext called on nil Response")
		return
	}
	if w == nil {
		log.Println("WARNING: SendWithContext called with nil ResponseWriter")
		return
	}

	if err := r.validateResponseSize(); err != nil {
		log.Printf("WARNING: Attempted to send response with invalid status code %d: %v. Response will be sent. as 500", r.Code, err)
		r.Code = 500
	}

	r.sendInternal(ctx, w)
}

// Internal send method to avoid code duplication
func (r *Response) sendInternal(ctx context.Context, w http.ResponseWriter) {
	if r == nil {
		log.Println("WARNING: sendInternal called on nil Response")
		return
	}
	if w == nil {
		log.Println("WARNING: sendInternal called with nil ResponseWriter")
		return
	}

	if err := validateStatusCode(r.Code); err != nil {
		log.Printf("WARNING: Attempted to send response with invalid status code %d: %v. Response will be sent. as 500", r.Code, err)
		r.Code = 500
	}

	interceptorsMu.RLock()
	currentInterceptors := make([]ResponseInterceptor, len(interceptors))
	copy(currentInterceptors, interceptors)
	interceptorsMu.RUnlock()

	for _, interceptor := range currentInterceptors {
		if interceptor != nil {
			if ctx != nil && ctx != context.Background() {
				interceptor.Intercept(ctx, r, r.Code)
			} else {
				interceptor.InterceptSimple(r, r.Code)
			}
		}
	}

	w.Header().Set("Content-Type", r.ContentType)
	w.WriteHeader(r.Code)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(r); err != nil {
		// If encoding fails, we can't send the original response so we leave it to Interceptors
		r.appendTraceInternal("internal error", (&EncodingError{Inner: err}).Error())
	}
}
