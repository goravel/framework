package session

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/errors"
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
		if c == nil {
			return nil, errors.ErrConfigFacadeNotSet.SetModule(errors.ModuleSession)
		}

		j := app.GetJson()
		if j == nil {
			return nil, errors.ErrJSONParserNotSet.SetModule(errors.ModuleSession)
		}

		return NewManager(c, j), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	SessionFacade = app.MakeSession()
	ConfigFacade = app.MakeConfig()
}
