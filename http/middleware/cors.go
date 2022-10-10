package middleware

import (
	"github.com/goravel/framework/contracts/http"
)

func Cors() http.Middleware {
	return func(ctx http.Context) {
		method := ctx.Request().Method()
		origin := ctx.Request().Header("Origin", "")
		if origin != "" {
			ctx.Response().Header("Access-Control-Allow-Origin", "*")
			ctx.Response().Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			ctx.Response().Header("Access-Control-Allow-Headers", "*")
			ctx.Response().Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Authorization")
			ctx.Response().Header("Access-Control-Max-Age", "172800")
			ctx.Response().Header("Access-Control-Allow-Credentials", "true")
		}

		if method == "OPTIONS" {
			ctx.Request().AbortWithStatus(204)
			return
		}

		ctx.Request().Next()
	}
}
