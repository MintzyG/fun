package response

import (
	"context"
	"log"
	"net/http"
)

// Error resolves err into an *AppError, builds a *Response, and sends it.
// It is the idiomatic one-liner for error paths in HTTP handlers:
//
//	user, err := svc.GetUser(ctx, id)
//	if err != nil {
//	    response.Error(err, w)
//	    return
//	}
//
// Resolution order (see resolveAppError for detail):
//  1. err is already an *AppError.
//  2. A mapper is registered via RegisterAppErrorMapper.
//  3. Fallback: log a warning, expose raw err.Error() as ErrInternal.
func Error(err error, w http.ResponseWriter) {
	if w == nil {
		log.Println("WARNING: response.Error called with nil ResponseWriter")
		return
	}
	ae := resolveAppError(err)
	ae.toResponse().Send(w)
}

// ErrorCtx is identical to Error but threads a context through to interceptors,
// matching the SendWithContext pattern used elsewhere in the library.
//
//	user, err := svc.GetUser(ctx, id)
//	if err != nil {
//	    response.ErrorCtx(ctx, err, w)
//	    return
//	}
func ErrorCtx(ctx context.Context, err error, w http.ResponseWriter) {
	if w == nil {
		log.Println("WARNING: response.ErrorCtx called with nil ResponseWriter")
		return
	}
	ae := resolveAppError(err)
	ae.toResponse().SendWithContext(ctx, w)
}
