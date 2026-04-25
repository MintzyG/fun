package handlers

import (
	"net/http"

	"github.com/MintzyG/FastUtilitiesNet"
)

// VersionHandler returns user-supplied build info as JSON.
// Populate the info map with whatever your service exposes — version string,
// commit hash, build time, etc. — typically injected via ldflags.
type VersionHandler struct {
	info map[string]any
}

// Version creates a version handler.
//
//	r.Get("/version", handlers.Version(map[string]any{
//	    "version":  version,
//	    "commit":   commit,
//	    "built_at": builtAt,
//	}).Handle)
func Version(info map[string]any) *VersionHandler {
	return &VersionHandler{info: info}
}

func (h *VersionHandler) Handle(w http.ResponseWriter, _ *http.Request) {
	fun.OK().
		WithContentType("application/json").
		WithData(h.info).
		Send(w)
}
