package translation

import (
	"context"
	"path/filepath"

	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (translation *ServiceProvider) Register(app foundation.Application) {
	app.BindWith(Binding, func(app foundation.Application, parameters map[string]any) (any, error) {
		config := app.MakeConfig()
		locale := config.GetString("app.locale")
		fallback := config.GetString("app.fallback_locale")
		loader := NewFileLoader([]string{filepath.Join("lang")})
		trans := NewTranslator(parameters["ctx"].(context.Context), loader, locale, fallback)
		trans.SetLocale(locale)
		return trans, nil
	})
}

func (translation *ServiceProvider) Boot(app foundation.Application) {

}
