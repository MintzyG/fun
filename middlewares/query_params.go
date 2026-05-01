package middlewares

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	fun "github.com/MintzyG/FastUtilitiesNet"
	"github.com/google/uuid"
)

type queryParamsKey[T any] struct{}

// QueryParams extracts the parsed query params from the request context.
func QueryParams[T any](r *http.Request) T {
	var zero T
	v := r.Context().Value(queryParamsKey[T]{})
	if v == nil {
		return zero
	}
	return v.(T)
}

func parseQueryTag(tag string) (key, defaultVal string, required bool) {
	parts := strings.Split(tag, ",")
	key = strings.TrimSpace(parts[0])
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "required" {
			required = true
		} else if strings.HasPrefix(part, "default=") {
			defaultVal = strings.TrimPrefix(part, "default=")
		}
	}
	return
}

// WithParams parses, converts, and validates query params into T using fun_query tags.
// Pass strict=true to reject any query param not declared in T.
//
// Tags:
//
//	fun_query:"key"               string param, optional
//	fun_query:"key,required"      param must be present and non-empty
//	fun_query:"key,default=val"   fallback value if absent or empty
//
// Supported field types: string, int, int64, bool, uuid.UUID
//
// Usage:
//
//	type ListParams struct {
//	    Status string    `fun_query:"status,default=active"`
//	    Page   int       `fun_query:"page,default=1"`
//	    ID     uuid.UUID `fun_query:"id,required"`
//	}
//
//	r.With(middlewares.WithParams[ListParams]()).Get("/items", h.List)
//	r.With(middlewares.WithParams[ListParams](true)).Get("/items", h.List) // strict
//
//	// in handler:
//	params := middlewares.Params[ListParams](r)
func WithParams[T any](strict ...bool) func(http.Handler) http.Handler {
	isStrict := len(strict) > 0 && strict[0]

	var zero T
	dstType := reflect.TypeOf(zero)

	// build allowed set and cache it once at middleware construction time
	allowed := make(map[string]struct{}, dstType.NumField())
	for i := 0; i < dstType.NumField(); i++ {
		tag := dstType.Field(i).Tag.Get("fun_query")
		if tag == "" {
			continue
		}
		key, _, _ := parseQueryTag(tag)
		allowed[key] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isStrict {
				for param := range r.URL.Query() {
					if _, ok := allowed[param]; !ok {
						fun.BadRequest("unknown query param: " + param).Send(w)
						return
					}
				}
			}

			var dst T
			req := fun.From(r)
			c := fun.Collect(req)

			dstVal := reflect.ValueOf(&dst).Elem()

			for i := 0; i < dstType.NumField(); i++ {
				field := dstType.Field(i)
				fieldVal := dstVal.Field(i)

				tag := field.Tag.Get("fun_query")
				if tag == "" {
					continue
				}

				key, defaultVal, required := parseQueryTag(tag)

				switch fieldVal.Interface().(type) {
				case string:
					if required {
						fieldVal.SetString(c.Query(key).String())
					} else {
						fieldVal.SetString(c.Query(key).StringOr(defaultVal))
					}

				case int:
					if required {
						fieldVal.SetInt(int64(c.Query(key).Int()))
					} else {
						def, _ := strconv.Atoi(defaultVal)
						fieldVal.SetInt(int64(c.Query(key).IntOr(def)))
					}

				case int64:
					if required {
						fieldVal.SetInt(c.Query(key).Int64())
					} else {
						def, _ := strconv.ParseInt(defaultVal, 10, 64)
						fieldVal.SetInt(c.Query(key).Int64Or(def))
					}

				case bool:
					if required {
						fieldVal.SetBool(c.Query(key).Bool())
					} else {
						fieldVal.SetBool(c.Query(key).BoolOr(defaultVal == "true"))
					}

				case uuid.UUID:
					id := c.Query(key).UUID()
					fieldVal.Set(reflect.ValueOf(id))
				}
			}

			if c.HasErrors() {
				fun.Error(c.Fail()).Send(w)
				return
			}

			ctx := context.WithValue(r.Context(), queryParamsKey[T]{}, dst)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
