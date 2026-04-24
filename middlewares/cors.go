package middlewares

import (
	"net/http"
	"time"

	"github.com/rs/cors"
)

// CORSConfig exposes the options you actually need to touch.
// Everything else is defaulted to safe, permissive-for-APIs values.
type CORSConfig struct {
	// AllowedOrigins is the list of origins allowed to make cross-origin requests.
	// Use []string{"*"} to allow all. Required.
	AllowedOrigins []string

	// AllowedMethods defaults to GET, POST, PUT, PATCH, DELETE, OPTIONS if empty.
	AllowedMethods []string

	// AllowedHeaders defaults to Content-Type, Authorization, X-Request-ID, X-API-Key if empty.
	AllowedHeaders []string

	// ExposedHeaders are headers the browser is allowed to read from the response.
	ExposedHeaders []string

	// AllowCredentials sets Access-Control-Allow-Credentials.
	// Note: cannot be used with AllowedOrigins: ["*"].
	AllowCredentials bool

	// MaxAge sets how long the preflight result can be cached.
	// Defaults to 10 minutes if zero.
	MaxAge time.Duration
}

// CORS returns a fully configured CORS middleware using rs/cors under the hood.
//
//	r.Use(mw.CORS(mw.CORSConfig{
//	    AllowedOrigins:   []string{"https://app.example.com"},
//	    AllowCredentials: true,
//	}))
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Request-ID", "X-API-Key"}
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 10 * time.Minute
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           int(cfg.MaxAge.Seconds()),
	})

	return c.Handler
}
