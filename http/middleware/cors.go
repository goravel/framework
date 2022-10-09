package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func Cors() http.Middleware {
	return func(request http.Request) {
		method := request.Method()
		origin := request.Header("Origin", "")
		if origin != "" {
			facades.Response.Header("Access-Control-Allow-Origin", "*")
			facades.Response.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			facades.Response.Header("Access-Control-Allow-Headers", "*")
			facades.Response.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Authorization")
			facades.Response.Header("Access-Control-Max-Age", "172800")
			facades.Response.Header("Access-Control-Allow-Credentials", "true")
		}

		if method == "OPTIONS" {
			request.AbortWithStatus(204)
			return
		}

		request.Next()
	}
}
