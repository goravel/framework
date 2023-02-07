package middleware

import (
	nethttp "net/http"

	"github.com/unrolled/secure"

	"github.com/goravel/framework/contracts/http"
)

func Tls(attr string) http.Middleware {
	return func(ctx http.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     attr,
		})

		if err := secureMiddleware.Process(ctx.Response().Writer(), ctx.Request().Origin()); err != nil {
			ctx.Request().AbortWithStatus(nethttp.StatusForbidden)
		}

		ctx.Request().Next()
	}
}
