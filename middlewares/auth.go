package middlewares

// Provides HTTP middleware for Bearer JWT and API key authentication.
// JWT verification is handled by the package; the caller supplies typed claims and hooks
// at construction time — methods are zero-argument at the call site.

import (
	"context"
	"net/http"
	"strings"

	"github.com/MintzyG/FastUtilitiesNet"
	"github.com/golang-jwt/jwt/v5"
)

// JWTHook is called after the package successfully verifies and parses a JWT.
// It may enrich ctx (e.g. inject a subject) and must return the (possibly new) ctx.
// Returning a non-nil error rejects the request with 401.
type JWTHook[C jwt.Claims] func(ctx context.Context, claims C) (context.Context, error)

// APIKeyHook is called with the raw key extracted from X-API-Key.
// Lookup, hashing, and ctx enrichment are fully the caller's responsibility.
// Returning a non-nil error rejects the request with 401.
type APIKeyHook func(ctx context.Context, rawKey string) (context.Context, error)

// KeyFunc returns the key used to verify a token, given its parsed (but unverified) form while also taking a context.
// Mirrors jwt.Keyfunc so callers can reuse existing implementations.
type KeyFunc[C jwt.Claims] func(ctx context.Context, tokenStr string) (C, error)

// Middleware holds all auth configuration. Build once with New, use everywhere.
type Middleware[C jwt.Claims] struct {
	KeyFunc    KeyFunc[C]
	jwtHook    JWTHook[C]
	apiKeyHook APIKeyHook
}

// New creates an auth middleware.
//   - keyFunc: provides the verification key (e.g. []byte HMAC secret, *rsa.PublicKey)
//   - jwtHook: called after successful JWT verification to enrich the ctx
//   - apiKeyHook: called with the raw API key — lookup and ctx enrichment are yours
//
// TracerName optionally overrides the OTel tracer name (defaults to "trimid/auth").
//
//	authMW := auth.New[*MyClaims](keyFunc, jwtHook, apiKeyHook)
//	r.With(authMW.JWT()).Get("/me", meHandler)
//	r.With(authMW.APIKey()).Post("/ingest", ingestHandler)
//	r.With(authMW.AnyAuth()).Post("/events", eventHandler)
func New[C jwt.Claims](keyFunc KeyFunc[C], jwtHook JWTHook[C], apiKeyHook APIKeyHook) *Middleware[C] {
	return &Middleware[C]{
		KeyFunc:    keyFunc,
		jwtHook:    jwtHook,
		apiKeyHook: apiKeyHook,
	}
}

// JWT returns a middleware that:
//  1. Extracts the Bearer token from Authorization.
//  2. Verifies and parses it into C using the configured KeyFunc.
//  3. Calls the configured JWTHook with the parsed claims.
//
// Any failure at any step responds 401 and stops the chain.
func (m *Middleware[C]) JWT() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tokenStr, ok := extractBearer(r)
			if !ok {
				writeUnauthorized(w, "missing or invalid Authorization header")
				return
			}

			claims, err := m.KeyFunc(ctx, tokenStr)
			if err != nil {
				writeUnauthorized(w, "invalid token")
				return
			}

			ctx, err = m.jwtHook(ctx, claims)
			if err != nil {
				writeUnauthorized(w, err.Error())
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// APIKey returns a middleware that:
//  1. Extracts the raw key from X-API-Key.
//  2. Calls the configured APIKeyHook — lookup, hashing, and ctx enrichment are yours.
//
// Any failure at any step responds 401 and stops the chain.
func (m *Middleware[C]) APIKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rawKey := r.Header.Get("X-API-Key")
			if rawKey == "" {
				writeUnauthorized(w, "missing X-API-Key header")
				return
			}

			ctx, err := m.apiKeyHook(ctx, rawKey)
			if err != nil {
				writeUnauthorized(w, err.Error())
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AnyAuth tries API key first (if X-API-Key is present), then Bearer JWT.
// At least one must succeed; if neither header is present the request is rejected.
func (m *Middleware[C]) AnyAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				m.APIKey()(next).ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
				m.JWT()(next).ServeHTTP(w, r)
				return
			}

			writeUnauthorized(w, "missing credentials")
		})
	}
}

func extractBearer(r *http.Request) (string, bool) {
	_, token, found := strings.Cut(r.Header.Get("Authorization"), "Bearer ")
	if !found || token == "" {
		return "", false
	}
	return token, true
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	fun.Unauthorized(msg).Send(w)
}
