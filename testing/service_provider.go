package testing

import (
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractsroute "github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

const Binding = "goravel.testing"

var artisanFacade contractsconsole.Artisan
var routeFacade contractsroute.Route
var configFacade contractsconfig.Config

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
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

	configFacade = app.MakeConfig()
	if configFacade == nil {
		color.Errorln(errors.ConfigFacadeNotSet.SetModule(errors.ModuleTesting))
	}
}
