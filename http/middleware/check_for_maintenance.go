package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func CheckForMaintenance() http.Middleware {
	return func(ctx http.Context) {
		if file.Exists(path.Storage("framework/down")) {
			ctx.Request().AbortWithStatus(http.StatusServiceUnavailable)
			return
		}

		ctx.Request().Next()
	}
}
