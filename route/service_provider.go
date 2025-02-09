package route

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (route *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingRoute, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleRoute)
		}

		return NewRoute(config)
	})
}

func (route *ServiceProvider) Boot(app foundation.Application) {

}
