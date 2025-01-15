package route

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.route"

type ServiceProvider struct {
}

func (route *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleRoute)
		}

		return NewRoute(config)
	})
}

func (route *ServiceProvider) Boot(app foundation.Application) {

}
