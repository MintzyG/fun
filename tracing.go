package fun

import (
	"fmt"
	"log"
)

// Takes in strings, errors and Stringers
func (r *Response) AddTrace(trace ...any) *Response {
	if r == nil {
		log.Println("WARNING: AddTrace called on nil Response")
		return nil
	}
	if r.TracePrefix == "" {
		return r.appendTrace("trace", false, trace...)
	}
	return r.appendTrace(r.TracePrefix, false, trace...)
}

// Takes in strings, errors and Stringers
func (r *Response) AddPrefixedTrace(prefix string, trace ...any) *Response {
	if r == nil {
		log.Println("WARNING: AddPrefixedTrace called on nil Response")
		return nil
	}
	if prefix == "" {
		return r.appendTrace("trace", false, trace...)
	}
	return r.appendTrace(prefix, false, trace...)
}

// AppendTraceInternal is for internal use and can override the last trace entry when full
func (r *Response) appendTraceInternal(prefix string, trace ...any) *Response {
	if r == nil {
		log.Println("WARNING: appendTraceInternal called on nil Response")
		return nil
	}
	return r.appendTrace(prefix, true, trace...)
}

// Internal trace appending logic
func (r *Response) appendTrace(prefix string, force bool, trace ...any) *Response {
	if r == nil {
		log.Println("WARNING: appendTrace called on nil Response")
		return nil
	}
	config := r.getResponseConfig()

	for _, t := range trace {
		var traceStr string
		switch v := t.(type) {
		case string:
			traceStr = v
		case error:
			traceStr = v.Error()
		case fmt.Stringer:
			traceStr = v.String()
		default:
			continue
		}

		traceStrFull := prefix + ": " + traceStr

		if len(r.Trace) < config.MaxTraceSize {
			r.Trace = append(r.Trace, traceStrFull)
		} else {
			if force && config.MaxTraceSize > 0 {
				r.Trace[config.MaxTraceSize-1] = traceStr
			} else if !force && config.MaxTraceSize > 0 {
				truncMsg := fmt.Sprintf("Error: (trace truncated, max size: %d)", config.MaxTraceSize)
				if r.Trace[config.MaxTraceSize-1] != truncMsg {
					r.Trace[config.MaxTraceSize-1] = truncMsg
				}
				break
			}
		}
	}
	return r
}
