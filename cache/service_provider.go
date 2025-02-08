package cache

import (
	"github.com/goravel/framework/cache/console"
	frameworkcontracts "github.com/goravel/framework/contracts"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (cache *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(frameworkcontracts.BindingCache, func(app foundation.Application) (any, error) {
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

func (cache *ServiceProvider) Boot(app foundation.Application) {
	cache.registerCommands(app)
}

func (cache *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		console.NewClearCommand(app.MakeCache()),
	})
}
