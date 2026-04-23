package route

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	routeconsole "github.com/goravel/framework/route/console"
	"github.com/goravel/framework/support/color"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Route,
		},
		Dependencies: binding.Bindings[binding.Route].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Route, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleRoute)
		}

		return NewRoute(config)
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	route := app.MakeRoute()
	if route == nil {
		color.Warningln(errors.RouteFacadeNotSet.Error())
		return
	}

	app.Commands([]console.Command{
		routeconsole.NewList(route),
	})
}

func (r *ServiceProvider) Runners(app foundation.Application) []foundation.Runner {
	return []foundation.Runner{NewRouteRunner(app.MakeConfig(), app.MakeRoute())}
}
