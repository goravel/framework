package middleware

import (
	nethttp "net/http"

	"github.com/unrolled/secure"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func Tls(host ...string) http.Middleware {
	return func(ctx http.Context) {
		if len(host) == 0 {
			defaultHost := facades.Config.GetString("route.tls.host")
			if defaultHost == "" {
				ctx.Request().Next()

				return
			}
			host = append(host, defaultHost)
		}

		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     host[0],
		})

		if err := secureMiddleware.Process(ctx.Response().Writer(), ctx.Request().Origin()); err != nil {
			ctx.Request().AbortWithStatus(nethttp.StatusForbidden)
		}

		ctx.Request().Next()
	}
}
