package auth

import (
	"context"

	"github.com/goravel/framework/auth/access"
	"github.com/goravel/framework/auth/console"
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	contractconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

var (
	cacheFacade  cache.Cache
	configFacade config.Config
	ormFacade    orm.Orm
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(contracts.BindingAuth, func(app foundation.Application, parameters map[string]any) (any, error) {
		configFacade = app.MakeConfig()
		if configFacade == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleAuth)
		}

		cacheFacade = app.MakeCache()
		if cacheFacade == nil {
			return nil, errors.CacheFacadeNotSet.SetModule(errors.ModuleAuth)
		}

		ormFacade = app.MakeOrm()
		if ormFacade == nil {
			// The Orm module will print the error message, so it's safe to return nil.
			return nil, nil
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleAuth)
		}

		ctx, ok := parameters["ctx"]
		if ok {
			return NewAuth(ctx.(http.Context), configFacade, log)
		}

		// ctx is unnecessary when calling facades.Auth().Extend()
		return NewAuth(nil, configFacade, log)
	})
	app.Singleton(contracts.BindingGate, func(app foundation.Application) (any, error) {
		return access.NewGate(context.Background()), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractconsole.Command{
		console.NewJwtSecretCommand(app.MakeConfig()),
		console.NewPolicyMakeCommand(),
	})
}
