package fun

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultKeepalive = 29 * time.Second

// SSE wraps a ResponseWriter for Server-Sent Events.
// Acquire one with NewSSE, then use Emit/Ping inside your select loop.
//
// Usage:
//
//	sse, err := fun.NewSSE(w, r)
//	if fun.Bail(w, err) { return }
//	defer sse.Close()
//
//	sse.Ping()
//	for {
//	    select {
//	    case <-sse.Done():
//	        return
//	    case <-sse.Keepalive():
//	        sse.Ping()
//	    case batch, ok := <-updates:
//	        if !ok { return }
//	        sse.Emit("inventory_update", batch)
//	    }
//	}
type SSE struct {
	w       http.ResponseWriter
	flusher http.Flusher
	ctx     context.Context
	ticker  *time.Ticker
}

// SSEOption configures an SSE instance.
type SSEOption func(*SSE)

// WithKeepalive sets the keepalive ping interval. Defaults to 29s.
func WithKeepalive(d time.Duration) SSEOption {
	return func(s *SSE) {
		s.ticker.Reset(d)
	}
}

// NewSSE sets SSE headers and returns a ready-to-use SSE writer.
// Returns an error if the ResponseWriter does not support flushing.
func NewSSE(w http.ResponseWriter, r *http.Request, opts ...SSEOption) (*SSE, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, ErrInternal("streaming not supported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	s := &SSE{
		w:       w,
		flusher: flusher,
		ctx:     r.Context(),
		ticker:  time.NewTicker(defaultKeepalive),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Emit encodes data as JSON and writes a named SSE event.
func (s *SSE) Emit(event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, payload)
	s.flusher.Flush()
	return err
}

// Ping writes an SSE comment to keep the connection alive.
func (s *SSE) Ping() {
	_, _ = fmt.Fprintf(s.w, ": ping\n\n")
	s.flusher.Flush()
}

// Done returns the request context's done channel.
func (s *SSE) Done() <-chan struct{} {
	return s.ctx.Done()
}

// Keepalive returns the ticker channel for periodic pings.
func (s *SSE) Keepalive() <-chan time.Time {
	return s.ticker.C
}

// Close stops the keepalive ticker. Call via defer after NewSSE.
func (s *SSE) Close() {
	s.ticker.Stop()
}
