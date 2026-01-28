package testing

import (
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractsroute "github.com/goravel/framework/contracts/route"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

var (
	json          foundation.Json
	artisanFacade contractsconsole.Artisan
	routeFacade   contractsroute.Route
	sessionFacade contractsession.Manager
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Testing,
		},
		Dependencies: binding.Bindings[binding.Testing].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Testing, func(app foundation.Application) (any, error) {
		artisan := app.MakeArtisan()
		if artisan == nil {
			return nil, errors.ConsoleFacadeNotSet.SetModule(errors.ModuleTesting)
		}
		cache := app.MakeCache()
		if cache == nil {
			return nil, errors.CacheFacadeNotSet.SetModule(errors.ModuleTesting)
		}
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleTesting)
		}
		orm := app.MakeOrm()
		if orm == nil {
			return nil, errors.OrmFacadeNotSet.SetModule(errors.ModuleTesting)
		}
		process := app.MakeProcess()
		if process == nil {
			return nil, errors.ProcessFacadeNotSet.SetModule(errors.ModuleTesting)
		}

		return NewApplication(artisan, cache, config, orm, process), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	artisanFacade = app.MakeArtisan()
	if artisanFacade == nil {
		color.Errorln(errors.ConsoleFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	routeFacade = app.MakeRoute()
	if routeFacade == nil {
		color.Errorln(errors.RouteFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	sessionFacade = app.MakeSession()
	if sessionFacade == nil {
		color.Errorln(errors.SessionFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	json = app.GetJson()
}
