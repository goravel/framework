package grpc

import (
	frameworkconfig "github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

func (route *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(frameworkconfig.BindingGrpc, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleGrpc)
		}

		return NewApplication(config), nil
	})
}

func (route *ServiceProvider) Boot(app foundation.Application) {
}
