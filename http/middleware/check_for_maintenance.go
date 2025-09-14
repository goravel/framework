package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
)

func CheckForMaintenance() http.Middleware {
	return func(ctx http.Context) {
		app := facades.App()
		if file.Exists(app.StoragePath("framework/down")) {
			ctx.Request().AbortWithStatus(http.StatusServiceUnavailable)
			return
		}

		ctx.Request().Next()
	}
}
