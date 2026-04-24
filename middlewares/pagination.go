package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
)

// PaginationType selects the pagination shape used by NewWithPagination.
type PaginationType int

const (
	PaginationOffset PaginationType = iota
	PaginationCursor
)

// PaginationConfig fully describes a pagination middleware.
// Use with NewWithPagination for custom shapes.
type PaginationConfig struct {
	Type PaginationType

	// Offset pagination params.
	PageParam    string // default: "page"
	LimitParam   string // default: "limit"
	DefaultLimit int    // default value for limit when absent
	MaxLimit     int    // requests above this are rejected

	// Cursor pagination params.
	CursorParam    string // default: "cursor"
	PerPageParam   string // default: "per_page"
	DefaultPerPage int
	MaxPerPage     int
}

func (c *PaginationConfig) applyDefaults() {
	if c.PageParam == "" {
		c.PageParam = "page"
	}
	if c.LimitParam == "" {
		c.LimitParam = "limit"
	}
	if c.CursorParam == "" {
		c.CursorParam = "cursor"
	}
	if c.PerPageParam == "" {
		c.PerPageParam = "per_page"
	}
}

// NewWithPagination returns a pre-baked pagination middleware from a full config.
// Use this when you need custom param names or a non-standard shape.
//
//	pagMW := mw.NewWithPagination(mw.PaginationConfig{
//	    Type:         mw.PaginationOffset,
//	    LimitParam:   "page_size",
//	    DefaultLimit: 20,
//	    MaxLimit:     200,
//	})
//	r.With(pagMW).Get("/items", handler)
func NewWithPagination(cfg PaginationConfig) func(http.Handler) http.Handler {
	cfg.applyDefaults()

	switch cfg.Type {
	case PaginationCursor:
		return buildCursorPagination(cfg.CursorParam, cfg.PerPageParam, cfg.DefaultPerPage, cfg.MaxPerPage)
	default:
		return buildOffsetPagination(cfg.PageParam, cfg.LimitParam, cfg.DefaultLimit, cfg.MaxLimit)
	}
}

// WithOffsetPagination is the predefined offset pagination middleware.
// It requires "page" and "limit", defaults limit to defaultLimit, and rejects
// values above maxLimit.
//
//	r.With(mw.WithOffsetPagination(20, 100)).Get("/items", handler)
func WithOffsetPagination(defaultLimit, maxLimit int) func(http.Handler) http.Handler {
	return buildOffsetPagination("page", "limit", defaultLimit, maxLimit)
}

// WithCursorPagination is the predefined cursor pagination middleware.
// It requires "cursor" and "per_page", defaults per_page to defaultPerPage, and
// rejects values above maxPerPage.
//
//	r.With(mw.WithCursorPagination(50, 200)).Get("/feed", handler)
func WithCursorPagination(defaultPerPage, maxPerPage int) func(http.Handler) http.Handler {
	return buildCursorPagination("cursor", "per_page", defaultPerPage, maxPerPage)
}

// buildOffsetPagination wires the primitives for offset-style pagination.
func buildOffsetPagination(pageParam, limitParam string, defaultLimit, maxLimit int) func(http.Handler) http.Handler {
	return chain(
		QueryAllow(pageParam, limitParam),
		QueryDefault(pageParam, "1"),
		QueryDefault(limitParam, strconv.Itoa(defaultLimit)),
		validateIntParam(limitParam, 1, maxLimit),
		validateIntParam(pageParam, 1, 0), // 0 max = no upper bound
	)
}

// buildCursorPagination wires the primitives for cursor-style pagination.
func buildCursorPagination(cursorParam, perPageParam string, defaultPerPage, maxPerPage int) func(http.Handler) http.Handler {
	return chain(
		QueryAllow(cursorParam, perPageParam),
		QueryRequire(cursorParam),
		QueryDefault(perPageParam, strconv.Itoa(defaultPerPage)),
		validateIntParam(perPageParam, 1, maxPerPage),
	)
}

// validateIntParam rejects requests where param is not a valid integer or falls
// outside [min, max]. Pass max=0 to skip the upper bound check.
func validateIntParam(param string, min, max int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.URL.Query().Get(param)
			if raw == "" {
				next.ServeHTTP(w, r)
				return
			}
			v, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(w, fmt.Sprintf("query param %q must be an integer", param), http.StatusBadRequest)
				return
			}
			if v < min {
				http.Error(w, fmt.Sprintf("query param %q must be >= %d", param, min), http.StatusBadRequest)
				return
			}
			if max > 0 && v > max {
				http.Error(w, fmt.Sprintf("query param %q must be <= %d", param, max), http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
