package fe

import (
	"runtime/debug"
	"sync/atomic"
)

var debugEnabled atomic.Bool

// SetDebug enables or disables debug extension output.
//
// When disabled (the default), [Builder.Debug] and [Builder.DebugErr] are no-ops.
// Enable in development/staging only — never in production, as stack traces
// and raw error messages can expose implementation internals (RFC 9457 §5).
//
//	errors.SetDebug(os.Getenv("ENV") != "production")
func SetDebug(enabled bool) {
	debugEnabled.Store(enabled)
}

// IsDebug reports whether debug output is currently enabled.
func IsDebug() bool {
	return debugEnabled.Load()
}

// debugInfo is the shape of the "debug" extension member.
type debugInfo struct {
	RawError   string `json:"raw_error,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// Debug attaches a "debug" extension to the Problem with raw error text and
// a captured stack trace. No-op when debug is disabled.
//
// The "debug" key is a custom extension; clients MUST ignore it if they
// don't recognize it (RFC 9457 §3.2).
//
//	errors.New(...).Debug(err).InternalServerError()
func (b *Builder) Debug(err error) *Builder {
	if !debugEnabled.Load() {
		return b
	}
	info := debugInfo{}
	if err != nil {
		info.RawError = err.Error()
	}
	info.StackTrace = string(debug.Stack())
	return b.With("debug", info)
}

// DebugMsg attaches a "debug" extension with a raw message string and a stack
// trace. No-op when debug is disabled.
func (b *Builder) DebugMsg(msg string) *Builder {
	if !debugEnabled.Load() {
		return b
	}
	return b.With("debug", debugInfo{
		RawError:   msg,
		StackTrace: string(debug.Stack()),
	})
}
