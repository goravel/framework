package log

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.log"

type ServiceProvider struct {
}

func (log *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleLog)
		}

		json := app.GetJson()
		if json == nil {
			return nil, errors.JSONParserNotSet.SetModule(errors.ModuleLog)
		}
		return NewApplication(config, json)
	})
}

func (log *ServiceProvider) Boot(app foundation.Application) {

}
