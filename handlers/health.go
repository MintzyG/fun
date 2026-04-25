package handlers

import (
	"net/http"

	"github.com/MintzyG/FastUtilitiesNet"
)

// HealthHandler is a liveness handler.
// Always returns 200 — if this endpoint is reachable, the process is alive.
type HealthHandler struct {
	service string
}

// Health creates a liveness handler for the given service name.
//
//	r.Get("/health", handlers.Health("my-api").Handle)
func Health(service string) *HealthHandler {
	return &HealthHandler{service: service}
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func (h *HealthHandler) Handle(w http.ResponseWriter, _ *http.Request) {
	payload := &healthResponse{
		Status:  "ok",
		Service: h.service,
	}

	fun.OK().WithData(payload).
		WithContentType("application/json").
		Send(w)
}
