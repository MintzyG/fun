package FUN

import (
	"errors"
	"log"
	"sync"
)

// AppErrorMapper converts any Go error into an *AppError.
// Register one at application startup via RegisterAppErrorMapper.
type AppErrorMapper func(err error) *AppError

var (
	appErrorMapper   AppErrorMapper
	appErrorMapperMu sync.RWMutex
)

// RegisterAppErrorMapper sets the global mapper used by Error to convert
// arbitrary errors into *AppError values. Call once during initialization.
func RegisterAppErrorMapper(m AppErrorMapper) {
	appErrorMapperMu.Lock()
	defer appErrorMapperMu.Unlock()
	appErrorMapper = m
}

// GetAppErrorMapper returns the currently registered mapper, or nil.
func GetAppErrorMapper() AppErrorMapper {
	appErrorMapperMu.RLock()
	defer appErrorMapperMu.RUnlock()
	return appErrorMapper
}

// ResetAppErrorMapper removes the registered mapper (useful in tests).
func ResetAppErrorMapper() {
	appErrorMapperMu.Lock()
	defer appErrorMapperMu.Unlock()
	appErrorMapper = nil
}

// Error resolves err into a *Response ready to be sent.
//
//	user, err := svc.GetUser(ctx, id)
//	if err != nil {
//	    response.Error(err).Send(w)
//	    return
//	}
//
// Resolution order:
//  1. err is already an *AppError — use it directly.
//  2. A mapper is registered — delegate to it.
//  3. No mapper — log a warning and wrap as ErrInternal.
func Error(err error) *Response {
	return resolveAppError(err).toResponse()
}

func resolveAppError(err error) *AppError {
	if err == nil {
		return &AppError{
			Code:    ErrInternal,
			Message: "an unknown error occurred",
		}
	}

	// 1. Already an AppError.
	if ae, ok := errors.AsType[*AppError](err); ok {
		return ae
	}

	// 2. Mapper registered.
	appErrorMapperMu.RLock()
	mapper := appErrorMapper
	appErrorMapperMu.RUnlock()

	if mapper != nil {
		if mapped := mapper(err); mapped != nil {
			return mapped
		}
	}

	// 3. No mapper — warn and expose raw message.
	log.Printf("WARNING: response.Error called with an unmapped error and no AppErrorMapper registered. "+
		"Register one via response.RegisterAppErrorMapper. Raw error: %v", err)

	return &AppError{
		Code:    ErrInternal,
		Message: err.Error(),
		Err:     err,
	}
}
