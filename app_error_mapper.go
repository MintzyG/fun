package fun

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
//	id, err := req.Path("id").UUID()
//	if err != nil {
//	    fun.Error(err).Send(w)
//	    return
//	}
//
// Resolution order:
//  1. err is already an *AppError — use it directly.
//  2. err is a *ParseError or *MissingParamError — mapped to a validation AppError.
//  3. err is a *BodyError — mapped to a bad request AppError.
//  4. A mapper is registered — delegate to it.
//  5. No mapper — log a warning and wrap as ErrInternal.
func Error(err error) *Response {
	return resolveAppError(err).toResponse()
}

func resolveAppError(err error) *AppError {
	if err == nil {
		return &AppError{
			Code:    CodeInternal,
			Message: "an unknown error occurred",
		}
	}

	// 1. Already an AppError.
	if ae, ok := errors.AsType[*AppError](err); ok {
		return ae
	}

	// 2. Param errors — map to validation AppError with field detail.
	if pe, ok := errors.AsType[*ParseError](err); ok {
		return Err("invalid params").
			WithFields(&FieldError{Field: pe.Src + "." + pe.Key, Message: err.Error(), Value: pe.Got}).
			Validation()
	}
	if me, ok := errors.AsType[*MissingParamError](err); ok {
		return Err("invalid params").
			WithFields(&FieldError{Field: me.Src + "." + me.Key, Message: err.Error()}).
			Validation()
	}

	// 3. Body decode error — map to bad request.
	if be, ok := errors.AsType[*BodyError](err); ok {
		return ErrBadRequest(be.Error())
	}

	// 4. Mapper registered.
	appErrorMapperMu.RLock()
	mapper := appErrorMapper
	appErrorMapperMu.RUnlock()

	if mapper != nil {
		if mapped := mapper(err); mapped != nil {
			return mapped
		}
	}

	// 5. No mapper — warn and expose raw message.
	log.Printf("WARNING: fun.Error called with an unmapped error and no AppErrorMapper registered. "+
		"Register one via fun.RegisterAppErrorMapper. Raw error: %v", err)

	return &AppError{
		Code:    CodeInternal,
		Message: err.Error(),
		Err:     err,
	}
}
