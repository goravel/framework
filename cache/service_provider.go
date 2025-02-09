package cache

import (
	"github.com/goravel/framework/cache/console"
	"github.com/goravel/framework/contracts"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingCache, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleCache)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleCache)
		}

		store := config.GetString("cache.default")

		return NewApplication(config, log, store)
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		console.NewClearCommand(app.MakeCache()),
	})
}
