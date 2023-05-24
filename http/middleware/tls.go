package middleware

import (
	"net/http"

	"github.com/unrolled/secure"

	httpcontract "github.com/goravel/framework/contracts/http"
	frameworkhttp "github.com/goravel/framework/http"
)

func Tls(host ...string) httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		if len(host) == 0 {
			defaultHost := frameworkhttp.ConfigFacade.GetString("http.tls.host")
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
