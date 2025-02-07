package hash

import (
	frameworkconfig "github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (hash *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(frameworkconfig.BindingHash, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleHash)
		}

		return NewApplication(config), nil
	})
}

func (hash *ServiceProvider) Boot(foundation.Application) {

}
