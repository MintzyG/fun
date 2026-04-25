package FUN

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
	AppError       *AppError       `json:"error,omitempty"`
	ContentType    string          `json:"-"`
	TracePrefix    string          `json:"-"`
	config         Config          `json:"-"`
	hasConfig      bool            `json:"-"`
}

func (r *Response) getResponseConfig() Config {
	if r.hasConfig {
		return r.config
	}
	return getConfig()
}

// WithConfig sets a custom configuration for this response instance,
// overriding the global configuration.
func (r *Response) WithConfig(config Config) *Response {
	if r == nil {
		log.Println("WARNING: WithConfig called on nil Response")
		return nil
	}
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
	r.hasConfig = true
	if r.ContentType == "" {
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

func (r *Response) WithCode(code int) *Response {
	if r == nil {
		log.Println("WARNING: WithCode called on nil Response")
		return nil
	}
	if err := validateStatusCode(code); err != nil {
		log.Printf("WARNING: WithCode called with invalid status code %d: %v", code, err)
		return r
	}
	r.Code = code
	return r
}

// Send writes the response to w. Use SendWithContext when a request context is available.
func (r *Response) Send(w http.ResponseWriter) {
	if r == nil {
		log.Println("WARNING: Send called on nil Response")
		return
	}
	if w == nil {
		log.Println("WARNING: Send called with nil ResponseWriter")
		return
	}
	r.sendInternal(context.Background(), w)
}

// SendWithContext writes the response to w, passing ctx to interceptors.
func (r *Response) SendWithContext(ctx context.Context, w http.ResponseWriter) {
	if r == nil {
		log.Println("WARNING: SendWithContext called on nil Response")
		return
	}
	if w == nil {
		log.Println("WARNING: SendWithContext called with nil ResponseWriter")
		return
	}
	r.sendInternal(ctx, w)
}

func (r *Response) sendInternal(ctx context.Context, w http.ResponseWriter) {
	if err := validateStatusCode(r.Code); err != nil {
		log.Printf("WARNING: invalid status code %d: %v. Defaulting to 500.", r.Code, err)
		r.Code = 500
	}

	if err := r.validateResponseSize(); err != nil {
		log.Printf("WARNING: response size validation failed: %v. Sending anyway.", err)
	}

	// Strip AppError.Debug in non-development environments.
	if r.AppError != nil && r.AppError.Debug != nil && !r.getResponseConfig().IsDevelopment {
		r.AppError.Debug = nil
	}

	interceptorsMu.RLock()
	current := make([]ResponseInterceptor, len(interceptors))
	copy(current, interceptors)
	interceptorsMu.RUnlock()

	for _, interceptor := range current {
		if interceptor == nil {
			continue
		}
		if ctx != context.Background() {
			interceptor.Intercept(ctx, r, r.Code)
		} else {
			interceptor.InterceptSimple(r, r.Code)
		}
	}

	w.Header().Set("Content-Type", r.ContentType)
	w.WriteHeader(r.Code)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(r); err != nil {
		log.Printf("WARNING: failed to encode response: %v", err)
	}
}
