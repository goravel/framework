package session

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/session"
)

var (
	SessionFacade session.Manager
	ConfigFacade  config.Config
)

const Binding = "goravel.session"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		c := app.MakeConfig()
		return NewManager(c), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	SessionFacade = app.MakeSession()
	ConfigFacade = app.MakeConfig()
}
