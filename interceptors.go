package fun

import (
	"context"
	"sync"
)

type ResponseInterceptor interface {
	// Intercept should be called when context is available
	Intercept(ctx context.Context, response *Response, statusCode int)

	// InterceptSimple should be called when no context is available
	InterceptSimple(response *Response, statusCode int)
}

// Thread-safe interceptors registry
var (
	interceptors   []ResponseInterceptor
	interceptorsMu sync.RWMutex
)

// AddInterceptor godoc
// Interceptors should only be added during downtimes or application initialization
func AddInterceptor(interceptor ResponseInterceptor) error {
	interceptorsMu.Lock()
	defer interceptorsMu.Unlock()

	config := getConfig()
	if len(interceptors) >= config.MaxInterceptorAmount {
		return &InterceptorLimitError{
			Current: len(interceptors),
			Max:     config.MaxInterceptorAmount,
		}
	}

	interceptors = append(interceptors, interceptor)
	return nil
}

func RemoveAllInterceptors() {
	interceptorsMu.Lock()
	defer interceptorsMu.Unlock()
	interceptors = nil
}

func GetInterceptors() []ResponseInterceptor {
	interceptorsMu.RLock()
	defer interceptorsMu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]ResponseInterceptor, len(interceptors))
	copy(result, interceptors)
	return result
}
