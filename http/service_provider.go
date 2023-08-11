package http

import (
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/http/console"
)

const Binding = "goravel.http"

type ServiceProvider struct {
}

func (http *ServiceProvider) Register(app foundation.Application) {
}

func (http *ServiceProvider) Boot(app foundation.Application) {
	http.registerCommands(app)
}

func (http *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]consolecontract.Command{
		&console.RequestMakeCommand{},
		&console.ControllerMakeCommand{},
		&console.MiddlewareMakeCommand{},
	})
}
