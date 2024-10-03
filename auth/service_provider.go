package auth

import (
	"context"

	"github.com/goravel/framework/auth/access"
	"github.com/goravel/framework/auth/console"
	contractconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

const BindingAuth = "goravel.auth"
const BindingGate = "goravel.gate"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(BindingAuth, func(app foundation.Application, parameters map[string]any) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ErrConfigFacadeNotSet.SetModule(errors.ModuleAuth)
		}
		cacheFacade := app.MakeCache()
		if cacheFacade == nil {
			return nil, errors.ErrCacheFacadeNotSet.SetModule(errors.ModuleAuth)
		}

		ormFacade := app.MakeOrm()
		if ormFacade == nil {
			return nil, errors.ErrOrmFacadeNotSet.SetModule(errors.ModuleAuth)
		}

		return NewAuth(config.GetString("auth.defaults.guard"),
			cacheFacade, config, parameters["ctx"].(http.Context), ormFacade), nil
	})
	app.Singleton(BindingGate, func(app foundation.Application) (any, error) {
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
