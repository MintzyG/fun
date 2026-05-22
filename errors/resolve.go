package fe

import (
	stderrors "errors"
	"log"
	"sync"
)

// Mapper converts an arbitrary Go error into a [*Problem].
//
// Return nil to fall through to the default internal error behavior.
// Register one at application startup via [RegisterMapper].
//
//	errors.RegisterMapper(func(err error) *errors.Problem {
//	    var notFound *myapp.NotFoundError
//	    if stderrors.As(err, &notFound) {
//	        return errors.New("https://api.example.com/problems/not-found").
//	            Title("Resource not found.").
//	            Detail(notFound.Error()).
//	            NotFound()
//	    }
//	    return nil
//	})
type Mapper func(err error) *Problem

var (
	globalMapper   Mapper
	globalMapperMu sync.RWMutex
)

// RegisterMapper sets the global mapper used by [Resolve] to convert
// arbitrary errors into [*Problem] values. Call once during initialization.
func RegisterMapper(m Mapper) {
	globalMapperMu.Lock()
	defer globalMapperMu.Unlock()
	globalMapper = m
}

// ResetMapper removes the registered mapper. Useful in tests.
func ResetMapper() {
	globalMapperMu.Lock()
	defer globalMapperMu.Unlock()
	globalMapper = nil
}

// Resolve converts any Go error into a [*Problem].
//
// Resolution order:
//  1. err is nil → generic 500 (always a caller bug; logged).
//  2. err is already a *Problem → returned as-is.
//  3. A [Mapper] is registered → delegate; use result if non-nil.
//  4. No mapper / mapper returned nil → log a warning, return generic 500.
//
// This function never panics and always returns a non-nil [*Problem].
func Resolve(err error) *Problem {
	if err == nil {
		log.Print("[fun/errors] WARNING: Resolve called with nil error")
		return internalProblem("an unknown error occurred")
	}

	// 1. Already a Problem.
	if p, ok := stderrors.AsType[*Problem](err); ok {
		return p
	}

	// 2. Registered mapper.
	globalMapperMu.RLock()
	mapper := globalMapper
	globalMapperMu.RUnlock()

	if mapper != nil {
		if mapped := mapper(err); mapped != nil {
			return mapped
		}
	}

	// 3. No mapper or mapper returned nil.
	log.Printf("[fun/errors] WARNING: Resolve called with unmapped error and no Mapper registered. "+
		"Register one via RegisterMapper. Raw error: %v", err)

	return internalProblem(err.Error())
}

func internalProblem(detail string) *Problem {
	return New(AboutBlank).
		Title("Internal Server Error").
		Detail(detail).
		InternalServerError()
}
