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
	cacheFacade = app.MakeCache()
	ormFacade = app.MakeOrm()

	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractconsole.Command{
		console.NewJwtSecretCommand(app.MakeConfig()),
		console.NewPolicyMakeCommand(),
	})
}
