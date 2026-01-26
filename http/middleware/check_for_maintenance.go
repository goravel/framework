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

func CheckForMaintenanceMode() http.Middleware {
	return func(ctx http.Context) {
		filepath := path.Storage("framework/maintenance")
		if !file.Exists(filepath) {
			ctx.Request().Next()
			return
		}

		content, err := os.ReadFile(filepath)

		if err != nil {
			ctx.Response().String(http.StatusServiceUnavailable, err.Error()).Abort()
			return
		}

		var maintenanceOptions *console.MaintenanceOptions
		err = json.Unmarshal(content, &maintenanceOptions)

		if err != nil {
			ctx.Response().String(http.StatusServiceUnavailable, err.Error()).Abort()
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

		if err = ctx.Response().String(maintenanceOptions.Status, maintenanceOptions.Reason).Abort(); err != nil {
			panic(err)
		}
	}
}
