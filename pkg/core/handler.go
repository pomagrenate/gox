package core

import (
	"errors"
	"net/http"
)

// Empty represents a request that has no body, query, or path parameters.
type Empty struct{}

// Handler wraps a strongly-typed function into an http.HandlerFunc.
// Req is the struct type that represents the incoming request.
// Res is the type returned by the handler on success.
func Handler[Req any, Res any](fn func(*Context, *Req) (*Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)

		var req Req

		// Determine if Req is not Empty
		// If it's not Empty, we bind and validate it
		if _, ok := any(req).(Empty); !ok {
			if err := BindAndValidate(r, &req); err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					ctx.JSON(http.StatusBadRequest, httpErr)
				} else {
					ctx.JSON(http.StatusBadRequest, &HTTPError{Message: err.Error()})
				}
				return
			}
		}

		res, err := fn(ctx, &req)
		if err != nil {
			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				ctx.JSON(http.StatusBadRequest, httpErr)
			} else {
				// Default to internal server error for unhandled errors
				ctx.JSON(http.StatusInternalServerError, &HTTPError{Message: err.Error()})
			}
			return
		}

		// If Res is not Empty, return it as JSON
		if res != nil {
			if _, ok := any(*res).(Empty); !ok {
				ctx.JSON(http.StatusOK, res)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
