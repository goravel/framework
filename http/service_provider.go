package http

import (
	"github.com/goravel/framework/contracts/config"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/http/console"
)

const Binding = "goravel.http"

var (
	ConfigFacade config.Config
)

type ServiceProvider struct {
}

func (http *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewContext(app.MakeConfig()), nil
	})
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewRateLimiter(), nil
	})
}

func (http *ServiceProvider) Boot(app foundation.Application) {
	ConfigFacade = app.MakeConfig()
	http.registerCommands(app)
}

func (http *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]consolecontract.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
