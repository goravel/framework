package translation

import (
	"context"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (translation *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(Binding, func(app foundation.Application, parameters map[string]any) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleLang)
		}

		logger := app.MakeLog()
		if logger == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleLang)
		}

		locale := config.GetString("app.locale")
		fallback := config.GetString("app.fallback_locale")
		path := config.GetString("app.lang_path", "lang")
		loader := NewFileLoader([]string{path}, app.GetJson())
		trans := NewTranslator(parameters["ctx"].(context.Context), loader, locale, fallback, logger)

		return trans, nil
	})
}

func (translation *ServiceProvider) Boot(app foundation.Application) {

}
