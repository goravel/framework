package http

import (
	nethttp "net/http"

	"github.com/goravel/framework/contracts/http"
)

// HTTPHandlerFuncToHandlerFunc converts a net/http HandlerFunc to a Goravel http.Handler
func HTTPHandlerFuncToHandlerFunc(h nethttp.HandlerFunc) http.Handler {
	return HTTPHandlerToHandler(h)
}

// HTTPHandlerToHandler converts a net/http Handler to a Goravel http.Handler
func HTTPHandlerToHandler(h nethttp.Handler) http.Handler {
	return http.HandlerFunc(func(ctx http.Context) http.Response {
		h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})
}

// HTTPMiddlewareToMiddleware converts a net/http middleware to a Goravel http.Middleware
func HTTPMiddlewareToMiddleware(mw func(nethttp.Handler) nethttp.Handler) http.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(ctx http.Context) http.Response {
			var response http.Response

			// create a net/http handler to wrap the Goravel handler
			nextHandler := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
				response = next.ServeHTTP(ctx)
			})

			// apply net/http handler
			mw(nextHandler).ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())

			return response
		})
	}
}

// HandlerToHTTPHandler converts a Goravel http.Handler to a net/http Handler
func HandlerToHTTPHandler(h http.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if err := h.ServeHTTP(NewContext(r, w)).Render(); err != nil {
			// Handle the error
			w.WriteHeader(nethttp.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	})
}
