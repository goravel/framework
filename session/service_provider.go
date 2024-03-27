package session

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/session"
)

var (
	Facade       session.Manager
	ConfigFacade config.Config
)

const Binding = "goravel.session"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		con := app.MakeConfig()
		return NewManager(con), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	Facade = app.MakeSession()
	ConfigFacade = app.MakeConfig()
}
