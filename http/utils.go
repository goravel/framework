package http

import (
	"github.com/goravel/framework/contracts/http"
	nethttp "net/http"
)

// ConvertHandlerFunc converts a net/http HandlerFunc to a Goravel http.Handler
func ConvertHandlerFunc(h nethttp.HandlerFunc) http.Handler {
	return ConvertHandler(h)
}

// ConvertHandler converts a net/http Handler to a Goravel http.Handler
func ConvertHandler(h nethttp.Handler) http.Handler {
	return http.HandlerFunc(func(ctx http.Context) http.Response {
		h.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
		return nil
	})
}

// ConvertMiddleware converts a net/http middleware to a Goravel http.Handler
func ConvertMiddleware(mw func(nethttp.Handler) nethttp.Handler) func(http.Handler) http.Handler {
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
