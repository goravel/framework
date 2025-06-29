package translation

import (
	"context"
	"io/fs"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	contractstranslation "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Lang,
		},
		Dependencies: []string{
			binding.Config,
			binding.Log,
		},
		ProvideFor: []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(contracts.BindingLang, func(app foundation.Application, parameters map[string]any) (any, error) {
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

		var loader contractstranslation.Loader
		if f, ok := config.Get("app.lang_fs").(fs.FS); ok {
			loader = NewFSLoader(path, f, app.GetJson())
		}

		return NewTranslator(parameters["ctx"].(context.Context), loader, NewFileLoader([]string{cast.ToString(path)}, app.GetJson()), locale, fallback, logger), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {

}
