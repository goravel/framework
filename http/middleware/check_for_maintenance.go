package middleware

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/file"
)

func CheckForMaintenance(app foundation.Application) http.Middleware {
	return func(ctx http.Context) {
		if file.Exists(app.StoragePath("framework/down")) {
			ctx.Request().AbortWithStatus(http.StatusServiceUnavailable)
			return
		}

		ctx.Request().Next()
	}
}
