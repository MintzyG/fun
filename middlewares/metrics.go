package middlewares

// Package metrics provides Prometheus HTTP instrumentation middleware.

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collectors holds the Prometheus metrics registered by this package.
// Use NewCollectors to create and register them, then pass to Middleware.
type Collectors struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

// NewCollectors creates and registers Prometheus metrics with the given registerer.
// Pass prometheus.DefaultRegisterer if you don't have a custom one.
func NewCollectors(reg prometheus.Registerer) (*Collectors, error) {
	c := &Collectors{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"route", "method", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of HTTP response times.",
				Buckets: prometheus.ExponentialBuckets(0.00025, 2, 16),
			},
			[]string{"route"},
		),
	}

	if err := reg.Register(c.RequestsTotal); err != nil {
		return nil, err
	}
	if err := reg.Register(c.RequestDuration); err != nil {
		return nil, err
	}
	return c, nil
}

// MetricsConfig controls which paths are skipped by the middleware.
type MetricsConfig struct {
	// SkipPrefixes lists URL path prefixes that will not be instrumented.
	// e.g. []string{"/metrics", "/swagger", "/healthz"}
	SkipPrefixes []string
}

// Metrics returns an HTTP middleware that records request count and duration.
func Metrics(c *Collectors, cfg MetricsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			skipRoute(w, r, next, cfg.SkipPrefixes)

			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			route := resolveRoutePattern(r)
			c.RequestsTotal.WithLabelValues(route, r.Method, http.StatusText(ww.status)).Inc()
			c.RequestDuration.WithLabelValues(route).Observe(time.Since(start).Seconds())
		})
	}
}

// Handler returns a standard Prometheus /metrics HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}
