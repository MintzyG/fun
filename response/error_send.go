package response

// Error resolves err into a *Response ready to be sent.
//
//	user, err := svc.GetUser(ctx, id)
//	if err != nil {
//	    response.Error(err).Send(w)
//	    return
//	}
//
// Resolution order (see resolveAppError for detail):
//  1. err is already an *AppError.
//  2. A mapper is registered via RegisterAppErrorMapper.
//  3. Fallback: log a warning, expose raw err.Error() as ErrInternal.
func Error(err error) *Response {
	return resolveAppError(err).toResponse()
}
