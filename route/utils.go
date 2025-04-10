package route

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
	return http.HandlerFunc(func(ctx http.Context) error {
		h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})
}

// HTTPMiddlewareToMiddleware converts a net/http middleware to a Goravel http.Middleware
func HTTPMiddlewareToMiddleware(mw func(nethttp.Handler) nethttp.Handler) http.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(ctx http.Context) error {
			var err error

			// create a net/http handler to wrap the Goravel handler
			nextHandler := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
				err = next.ServeHTTP(ctx)
			})

			// apply net/http handler
			mw(nextHandler).ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())

			return err
		})
	}
}
