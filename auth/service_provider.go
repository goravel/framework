package auth

import (
	"context"

	"github.com/goravel/framework/auth/access"
	"github.com/goravel/framework/auth/console"
	frameworkconfig "github.com/goravel/framework/config"
	contractconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(frameworkconfig.BindingAuth, func(app foundation.Application, parameters map[string]any) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleAuth)
		}
		cacheFacade := app.MakeCache()
		if cacheFacade == nil {
			return nil, errors.CacheFacadeNotSet.SetModule(errors.ModuleAuth)
		}

		ormFacade := app.MakeOrm()
		if ormFacade == nil {
			// The Orm module will print the error message, so it's safe to return nil.
			return nil, nil
		}

		ctx, ok := parameters["ctx"].(http.Context)
		if !ok {
			return nil, errors.InvalidHttpContext.SetModule(errors.ModuleAuth)
		}

		return NewAuth(config.GetString("auth.defaults.guard"),
			cacheFacade, config, ctx, ormFacade), nil
	})
	app.Singleton(frameworkconfig.BindingGate, func(app foundation.Application) (any, error) {
		return access.NewGate(context.Background()), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	database.registerCommands(app)
}

func (database *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractconsole.Command{
		console.NewJwtSecretCommand(app.MakeConfig()),
		console.NewPolicyMakeCommand(),
	})
}
