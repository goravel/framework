package middleware

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/foundation/console"
)

func CheckForMaintenanceMode() http.Middleware {
	return func(ctx http.Context) {
		storage := facades.Storage()
		filepath := "framework/maintenance.json"
		if !storage.Exists(filepath) {
			ctx.Request().Next()
			return
		}

		content, err := storage.GetBytes(filepath)

		if err != nil {
			if err = ctx.Response().String(http.StatusServiceUnavailable, err.Error()).Abort(); err != nil {
				panic(err)
			}
			return
		}

		var maintenanceOptions *console.MaintenanceOptions
		err = json.Unmarshal(content, &maintenanceOptions)

		if err != nil {
			if err = ctx.Response().String(http.StatusServiceUnavailable, err.Error()).Abort(); err != nil {
				panic(err)
			}
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
			if err = ctx.Response().View().Make(maintenanceOptions.Render).Render(); err != nil {
				panic(err)
			}
			return
		}

		if err = ctx.Response().String(maintenanceOptions.Status, maintenanceOptions.Reason).Abort(); err != nil {
			panic(err)
		}
	}
}
