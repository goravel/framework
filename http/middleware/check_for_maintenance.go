package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func CheckForMaintenance() http.Middleware {
	return func(ctx http.Context) {
		filepath := path.Storage("framework/maintenance")
		if file.Exists(filepath) {
			content, err := file.GetContent(filepath)

			if err != nil {
				ctx.Request().Abort(http.StatusServiceUnavailable)
				return
			}

			// Checking err to suppress the linter
			if err = ctx.Response().String(http.StatusServiceUnavailable, content).Abort(); err != nil {
				return
			}
			return
		}

		ctx.Request().Next()
	}
}
