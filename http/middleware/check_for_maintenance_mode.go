package middleware

import (
	"encoding/json"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/http"
)

func CheckForMaintenanceMode() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		maintenance := console.NewMaintenanceMode(http.App.MakeConfig(), http.App.MakeCache(), http.App.MakeStorage())
		content, exists, err := maintenance.Get()
		if err != nil {
			abortMaintenanceMode(ctx, err)
			return
		}
		if !exists {
			ctx.Request().Next()
			return
		}

		var maintenanceOptions console.MaintenanceOptions
		err = json.Unmarshal(content, &maintenanceOptions)

		if err != nil {
			abortMaintenanceMode(ctx, err)
			return
		}

		secret := ctx.Request().Query("secret", "")
		if secret != "" && maintenanceOptions.Secret != "" {
			hash := http.App.MakeHash()
			if hash == nil {
				panic(errors.HashFacadeNotSet)
			}

			if hash.Check(secret, maintenanceOptions.Secret) {
				ctx.Request().Next()
				return
			}
		}

		if maintenanceOptions.Redirect != "" {
			if ctx.Request().Path() == maintenanceOptions.Redirect {
				ctx.Request().Next()
				return
			}

			if err = ctx.Response().Redirect(httpcontract.StatusTemporaryRedirect, maintenanceOptions.Redirect).Abort(); err != nil {
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

func abortMaintenanceMode(ctx httpcontract.Context, err error) {
	if err = ctx.Response().String(httpcontract.StatusServiceUnavailable, err.Error()).Abort(); err != nil {
		panic(err)
	}
}
