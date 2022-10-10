package middleware

import (
	"fmt"
	"github.com/goravel/framework/contracts/http"
)

func Logger() http.Middleware {
	return func(ctx http.Context) {
		fmt.Println("hwb---------", ctx.Request().FullUrl())

		ctx.Request().Next()
	}
}
