package hash

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (hash *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingHash, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleHash)
		}

		return NewApplication(config), nil
	})
}

func (hash *ServiceProvider) Boot(foundation.Application) {

}
