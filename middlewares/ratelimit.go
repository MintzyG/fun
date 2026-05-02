package middlewares

import (
	"net/http"
	"sync"

	"github.com/MintzyG/fun"
	"golang.org/x/time/rate"
)

// KeyExtractorFunc extracts a rate limit key from the request.
// The returned string is used to bucket requests — typically an IP or API key.
type KeyExtractorFunc func(r *http.Request) string

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	// RPS is the number of requests per second allowed per key.
	RPS rate.Limit

	// Burst is the maximum burst size above RPS.
	Burst int

	// KeyExtractor determines how to bucket requests.
	// Defaults to RemoteAddr (IP-based) if nil.
	KeyExtractor KeyExtractorFunc
}

type rateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	cfg      RateLimitConfig
}

func (rl *rateLimiter) get(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	l, ok := rl.limiters[key]
	if !ok {
		l = rate.NewLimiter(rl.cfg.RPS, rl.cfg.Burst)
		rl.limiters[key] = l
	}
	return l
}

// RateLimit returns a token-bucket rate limiting middleware.
// Each unique key (by default, the client IP) gets its own limiter.
// Requests that exceed the limit receive 429 Too Many Requests.
//
//	// 100 req/s per IP, burst of 20
//	r.Use(mw.RateLimit(mw.RateLimitConfig{RPS: 100, Burst: 20}))
//
//	// per API key
//	r.Use(mw.RateLimit(mw.RateLimitConfig{
//	    RPS:   50,
//	    Burst: 10,
//	    KeyExtractor: func(r *http.Request) string {
//	        return r.Header.Get("X-API-Key")
//	    },
//	}))
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.KeyExtractor == nil {
		cfg.KeyExtractor = remoteIP
	}

	rl := &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		cfg:      cfg,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.KeyExtractor(r)
			if !rl.get(key).Allow() {
				fun.TooManyRequests("rate limit exceeded").Send(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// remoteIP extracts the IP portion of RemoteAddr.
func remoteIP(r *http.Request) string {
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
