package crypt

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.crypt"

type ServiceProvider struct {
}

func (crypt *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleCrypt)
		}

		json := app.GetJson()
		if json == nil {
			return nil, errors.JSONParserNotSet.SetModule(errors.ModuleCrypt)
		}

		return NewAES(config, json)
	})
}

func (crypt *ServiceProvider) Boot(app foundation.Application) {

}
