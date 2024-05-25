package translation

import (
	"context"
	"path/filepath"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/path"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (translation *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(Binding, func(app foundation.Application, parameters map[string]any) (any, error) {
		executable, err := path.Executable()
		if err != nil {
			return nil, err
		}

		config := app.MakeConfig()
		logger := app.MakeLog()
		locale := config.GetString("app.locale")
		fallback := config.GetString("app.fallback_locale")
		loader := NewFileLoader([]string{filepath.Join(executable, "lang")})
		trans := NewTranslator(parameters["ctx"].(context.Context), loader, locale, fallback, logger)
		trans.SetLocale(locale)
		return trans, nil
	})
}

func (translation *ServiceProvider) Boot(app foundation.Application) {

}
