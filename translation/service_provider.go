package translation

import (
	"context"
	"io/fs"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/foundation"
	contractstranslation "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(contracts.BindingTranslation, func(app foundation.Application, parameters map[string]any) (any, error) {
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
		path := config.Get("app.lang_path", "lang")

		var loader contractstranslation.Loader
		if f, ok := path.(fs.FS); ok {
			loader = NewFSLoader(f, app.GetJson())
		} else {
			loader = NewFileLoader([]string{cast.ToString(path)}, app.GetJson())
		}

		return NewTranslator(parameters["ctx"].(context.Context), loader, locale, fallback, logger), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {

}
