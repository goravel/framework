package testing

import (
	"github.com/goravel/framework/config"
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

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(config.BindingTesting, func(app foundation.Application) (any, error) {
		return NewApplication(app), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	artisanFacade = app.MakeArtisan()
	if artisanFacade == nil {
		color.Errorln(errors.ArtisanFacadeNotSet.SetModule(errors.ModuleTesting))
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
