package middleware

import (
	"encoding/json"
	"os"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
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

		if err != nil {
			ctx.Request().Abort(http.StatusServiceUnavailable)
			return
		}

		var maintenanceOptions *console.MaintenanceOptions
		err = json.Unmarshal(content, &maintenanceOptions)

		if err != nil {
			ctx.Request().Abort(http.StatusServiceUnavailable)
			return
		}

		secret := ctx.Request().Query("secret", "")
		if secret != "" && maintenanceOptions.Secret != "" {
			if facades.Hash().Check(secret, maintenanceOptions.Secret) {
				ctx.Request().Next()
				return
			}
		}

		if maintenanceOptions.Redirect != "" {
			if ctx.Request().Path() == maintenanceOptions.Redirect {
				ctx.Request().Next()
				return
			}

			if err = ctx.Response().Redirect(http.StatusTemporaryRedirect, maintenanceOptions.Redirect).Abort(); err != nil {
				return
			}
			return
		}

		if maintenanceOptions.Render != "" {
			ctx.Request().Abort(maintenanceOptions.Status)
			if err = ctx.Response().View().Make(maintenanceOptions.Render, map[string]string{}).Render(); err != nil {
				return
			}
			return
		}

		// Checking err to suppress the linter
		if err = ctx.Response().String(maintenanceOptions.Status, maintenanceOptions.Reason).Abort(); err != nil {
			return
		}
	}
}
