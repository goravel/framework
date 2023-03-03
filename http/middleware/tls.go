package middleware

import (
	"net/http"

	"github.com/unrolled/secure"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func Tls(host ...string) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		if len(host) == 0 {
			defaultHost := facades.Config.GetString("http.tls.host")
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
			ctx.Request().AbortWithStatus(http.StatusForbidden)
		}

		ctx.Request().Next()
	}
}
