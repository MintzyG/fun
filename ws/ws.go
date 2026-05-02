// Package ws wraps gorilla/websocket with a typed, low-boilerplate API
// consistent with the fun library's style.
//
// Usage:
//
//	conn, err := ws.Upgrade(w, r)
//	if fun.Bail(w, err) { return }
//	defer conn.Close()
//
//	msg, err := ws.Read[BuyRequest](conn)
//	if err != nil { return err }
//	// msg.Type, msg.Payload already typed — no double marshal
//
//	conn.Write("reservation_confirmed", ReservationConfirmedPayload{...})
//	conn.WriteError("invalid payload")
package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Conn wraps a gorilla websocket connection.
type Conn struct {
	conn *websocket.Conn
}

// RawMessage is the wire format: {"type": "...", "payload": ...}
type RawMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Message is a typed inbound message.
type Message[T any] struct {
	Type    string
	Payload T
}

// UpgradeOption configures the upgrader.
type UpgradeOption func(*websocket.Upgrader)

// WithOriginCheck sets a custom origin check function.
func WithOriginCheck(fn func(r *http.Request) bool) UpgradeOption {
	return func(u *websocket.Upgrader) {
		u.CheckOrigin = fn
	}
}

// WithBufferSize sets read and write buffer sizes.
func WithBufferSize(read, write int) UpgradeOption {
	return func(u *websocket.Upgrader) {
		u.ReadBufferSize = read
		u.WriteBufferSize = write
	}
}

// Upgrade performs the WebSocket handshake and returns a Conn.
// Returns an error if the upgrade fails — use fun.Bail(w, err) to handle it.
//
//	conn, err := ws.Upgrade(w, r)
//	if fun.Bail(w, err) { return }
//	defer conn.Close()
func Upgrade(w http.ResponseWriter, r *http.Request, opts ...UpgradeOption) (*Conn, error) {
	u := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	for _, opt := range opts {
		opt(u)
	}
	c, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("ws: upgrade failed: %w", err)
	}
	return &Conn{conn: c}, nil
}

// Read reads the next message and decodes the payload into T.
// Eliminates the double marshal/unmarshal pattern.
//
//	msg, err := ws.Read[BuyRequest](conn)
//	if err != nil { return err }
func Read[T any](c *Conn) (Message[T], error) {
	var raw RawMessage
	if err := c.conn.ReadJSON(&raw); err != nil {
		return Message[T]{}, err
	}
	var payload T
	if len(raw.Payload) > 0 && string(raw.Payload) != "null" {
		if err := json.Unmarshal(raw.Payload, &payload); err != nil {
			return Message[T]{}, fmt.Errorf("ws: failed to decode payload for type %q: %w", raw.Type, err)
		}
	}
	return Message[T]{Type: raw.Type, Payload: payload}, nil
}

// Write sends a typed message to the client.
//
//	conn.Write("reservation_confirmed", ReservationConfirmedPayload{...})
func (c *Conn) Write(msgType string, payload any) error {
	return c.conn.WriteJSON(RawMessage{
		Type:    msgType,
		Payload: MustMarshal(payload),
	})
}

// WriteRaw sends a pre-built RawMessage — useful when forwarding webhook messages.
func (c *Conn) WriteRaw(msg RawMessage) error {
	return c.conn.WriteJSON(msg)
}

// WriteError sends an error message to the client.
func (c *Conn) WriteError(reason string) error {
	return c.Write("error", reason)
}

// SetReadDeadline sets the deadline for the next read.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// Raw returns the underlying gorilla connection for advanced use.
func (c *Conn) Raw() *websocket.Conn {
	return c.conn
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// MustMarshal marshals v to json.RawMessage, returning null on failure.
func MustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return b
}
