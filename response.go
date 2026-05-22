package fun

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Status         int             `json:"status,omitempty"`
	Title          string          `json:"title,omitempty"`
	Type           ErrorCode       `json:"type,omitempty"`
	Detail         string          `json:"detail,omitempty"`
	Instance       string          `json:"instance,omitempty"`
	Data           any             `json:"data,omitempty"`
	Fields         []FieldError    `json:"fields,omitempty"`
	Meta           map[string]any  `json:"meta,omitempty"`
	PaginationData *PaginationMeta `json:"pagination,omitempty"`
	Timestamp      time.Time       `json:"timestamp,omitempty"`
	Module         string          `json:"module,omitempty"`
	ContentType    string          `json:"-"`
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
		log.Println("[fun] WARNING: WithConfig called on nil Response")
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
		log.Println("[fun] WARNING: WithContentType called on nil Response")
		return nil
	}
	r.ContentType = ctype
	return r
}

func (r *Response) WithModule(module string) *Response {
	if r == nil {
		log.Println("[fun] WARNING: WithModule called on nil Response")
		return nil
	}
	r.Module = module
	return r
}

func (r *Response) WithTitle(message string) *Response {
	if r == nil {
		log.Println("[fun] WARNING: WithMsg called on nil Response")
		return nil
	}
	r.Title = message
	return r
}

func (r *Response) WithData(data any) *Response {
	if r == nil {
		log.Println("[fun] WARNING: WithData called on nil Response")
		return nil
	}
	r.Data = data
	return r
}

func (r *Response) WithStatus(code int) *Response {
	if r == nil {
		log.Println("[fun] WARNING: WithCode called on nil Response")
		return nil
	}
	if err := validateStatusCode(code); err != nil {
		log.Printf("[fun] WARNING: WithCode called with invalid status code %d: %v", code, err)
		return r
	}
	r.Status = code
	return r
}

// Send writes the response to w. Use SendWithCtx when a request context is available.
func (r *Response) Send(w http.ResponseWriter) {
	if r == nil {
		log.Println("[fun] WARNING: Send called on nil Response")
		return
	}
	if w == nil {
		log.Println("[fun] WARNING: Send called with nil ResponseWriter")
		return
	}
	r.sendInternal(context.Background(), w)
}

// SendWithCtx writes the response to w, passing ctx to interceptors.
func (r *Response) SendWithCtx(ctx context.Context, w http.ResponseWriter) {
	if r == nil {
		log.Println("[fun] WARNING: SendWithContext called on nil Response")
		return
	}
	if w == nil {
		log.Println("[fun] WARNING: SendWithContext called with nil ResponseWriter")
		return
	}
	r.sendInternal(ctx, w)
}

func (r *Response) sendInternal(ctx context.Context, w http.ResponseWriter) {
	if err := validateStatusCode(r.Status); err != nil {
		log.Printf("[fun] WARNING: invalid status code %d: %v. Defaulting to 500.", r.Status, err)
		r.Status = 500
	}

	if err := r.validateResponseSize(); err != nil {
		log.Printf("[fun] WARNING: response size validation failed: %v. Sending anyway.", err)
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
			interceptor.Intercept(ctx, r, r.Status)
		} else {
			interceptor.InterceptSimple(r, r.Status)
		}
	}

	w.Header().Set("Content-Type", r.ContentType)
	w.WriteHeader(r.Status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(r); err != nil {
		log.Printf("[fun] WARNING: failed to encode response: %v", err)
	}
}
