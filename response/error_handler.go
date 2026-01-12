package response

import (
	"sync"
)

// ErrorHandler is a function type that converts an error into a Response
type ErrorHandler func(err error) *Response

// Default error handler
var (
	errorHandler   ErrorHandler
	errorHandlerMu sync.RWMutex
)

// defaultErrorHandler provides a basic error-to-response conversion
func defaultErrorHandler(err error) *Response {
	if err == nil {
		return InternalServerError("Unknown error occurred")
	}
	return InternalServerError(err.Error()).AddTrace(err)
}

func init() {
	errorHandler = defaultErrorHandler
}

// RegisterErrorHandler sets a custom error handler function
// This function will be used by FromError to convert errors into responses
// Note: It's recommended to set this via Config.ErrorHandler instead
func RegisterErrorHandler(handler ErrorHandler) {
	if handler == nil {
		handler = defaultErrorHandler
	}

	errorHandlerMu.Lock()
	defer errorHandlerMu.Unlock()
	errorHandler = handler
}

// GetErrorHandler returns the currently registered error handler
func GetErrorHandler() ErrorHandler {
	errorHandlerMu.RLock()
	defer errorHandlerMu.RUnlock()
	return errorHandler
}

// ResetErrorHandler resets the error handler to the default implementation
func ResetErrorHandler() {
	errorHandlerMu.Lock()
	defer errorHandlerMu.Unlock()
	errorHandler = defaultErrorHandler
}

// FromError converts an error into a Response using the registered error handler
func FromError(err error) *Response {
	errorHandlerMu.RLock()
	handler := errorHandler
	errorHandlerMu.RUnlock()

	return handler(err)
}
