package middleware

import (
	"encoding/json"
	"os"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func CheckForMaintenance() http.Middleware {
	return func(ctx http.Context) {
		filepath := path.Storage("framework/maintenance")
		if !file.Exists(filepath) {
			ctx.Request().Next()
		}

		content, err := os.ReadFile(filepath)

		var maintenanceOptions *console.MaintenanceOptions
		err = json.Unmarshal(content, &maintenanceOptions)

		if err != nil {
			ctx.Request().Abort(http.StatusServiceUnavailable)
			return
		}

		secret := ctx.Request().Query("secret", "")
		if secret != "" && maintenanceOptions.Secret != "" && secret == maintenanceOptions.Secret {
			ctx.Request().Next()
			return
		}

		if maintenanceOptions.Redirect != "" {
			ctx.Response().Redirect(http.StatusTemporaryRedirect, maintenanceOptions.Redirect)
			return
		}

		if maintenanceOptions.Render != "" {
			ctx.Response().View().Make(maintenanceOptions.Render, nil).Render()
			return
		}

		// Checking err to suppress the linter
		if err = ctx.Response().String(maintenanceOptions.Status, maintenanceOptions.Reason).Abort(); err != nil {
			return
		}
	}
}
