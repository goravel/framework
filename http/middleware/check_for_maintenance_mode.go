package middleware

import (
	"encoding/json"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/http"
)

type MaintenanceMode struct{}

func (m *MaintenanceMode) Signature() string {
	return "check_for_maintenance_mode"
}

func (m *MaintenanceMode) Handle(ctx contractshttp.Context) {
	config := http.App.MakeConfig()
	cache := http.App.MakeCache()
	storage := http.App.MakeStorage()
	hash := http.App.MakeHash()
	if config == nil || cache == nil || storage == nil || hash == nil {
		ctx.Request().Next()
		return
	}

	maintenance := console.NewMaintenanceMode(config, cache, storage)
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

		if err = ctx.Response().Redirect(contractshttp.StatusTemporaryRedirect, maintenanceOptions.Redirect).Abort(); err != nil {
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

func CheckForMaintenanceMode() contractshttp.Middleware {
	return &MaintenanceMode{}
}

func abortMaintenanceMode(ctx contractshttp.Context, err error) {
	if err = ctx.Response().String(contractshttp.StatusServiceUnavailable, err.Error()).Abort(); err != nil {
		panic(err)
	}
}
