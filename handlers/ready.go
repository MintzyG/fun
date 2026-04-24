package handlers

import (
	"context"
	"net/http"
	"sync"

	"github.com/MintzyG/FastUtilitiesNet/response"
)

// CheckFunc is a named readiness check.
// Return a non-nil error if the dependency is unavailable.
type CheckFunc struct {
	Name  string
	Check func(ctx context.Context) error
}

// ReadyHandler is a readiness handler.
// Runs all registered checks in parallel and reports per-check status.
type ReadyHandler struct {
	service string
	checks  []CheckFunc
}

// Ready creates a readiness handler.
//
//	r.Get("/ready", handlers.Ready("my-api",
//	    handlers.CheckFunc{Name: "postgres", Check: db.Ping},
//	    handlers.CheckFunc{Name: "redis",    Check: cache.Ping},
//	).Handle)
func Ready(service string, checks ...CheckFunc) *ReadyHandler {
	return &ReadyHandler{service: service, checks: checks}
}

type readyResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks"`
}

// Handle runs all checks in parallel.
// Returns 200 if all pass, 503 if any fail.
// The checks map always contains every check name — value is "ok" or the error message.
func (h *ReadyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	results := make(map[string]string, len(h.checks))
	healthy := true

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, c := range h.checks {
		wg.Add(1)
		go func(c CheckFunc) {
			defer wg.Done()
			err := c.Check(r.Context())
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[c.Name] = err.Error()
				healthy = false
			} else {
				results[c.Name] = "ok"
			}
		}(c)
	}
	wg.Wait()

	payload := readyResponse{
		Status:  "healthy",
		Service: h.service,
		Checks:  results,
	}

	if !healthy {
		payload.Status = "unhealthy"
		response.ServiceUnavailable("unhealthy").WithData(payload).
			WithContentType("application/json").
			Send(w)
		return
	}

	response.OK("healthy").WithData(payload).
		WithContentType("application/json").
		Send(w)
}
