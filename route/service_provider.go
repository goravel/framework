package route

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	routeconsole "github.com/goravel/framework/route/console"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingRoute, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleRoute)
		}

		return NewRoute(config)
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]console.Command{
		routeconsole.NewList(app.MakeRoute()),
	})
}
