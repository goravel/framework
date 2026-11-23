package testing

import (
	contractsbinding "github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractsroute "github.com/goravel/framework/contracts/route"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/binding"
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

func (r *ServiceProvider) Relationship() contractsbinding.Relationship {
	bindings := []string{
		contractsbinding.Testing,
		contractsbinding.Fake,
	}

	return contractsbinding.Relationship{
		Bindings:     bindings,
		Dependencies: binding.Dependencies(bindings...),
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contractsbinding.Testing, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleTesting)
		}

		// They are checked when using actually
		artisan := app.MakeArtisan()
		cache := app.MakeCache()
		orm := app.MakeOrm()
		process := app.MakeProcess()

		return NewApplication(artisan, cache, config, orm, process), nil
	})

	app.Singleton(contractsbinding.Fake, func(app foundation.Application) (any, error) {
		return NewFake(), nil
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
