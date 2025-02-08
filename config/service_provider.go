package config

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support"
)

type ServiceProvider struct {
}

func (config *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return NewApplication(support.EnvPath), nil
	})
}

func (config *ServiceProvider) Boot(app foundation.Application) {

}
