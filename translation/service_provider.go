package translation

import (
	"path/filepath"

	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (translation *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		local := config.GetString("app.locale")
		fallback := config.GetString("app.fallback_locale")
		loader := NewFileLoader([]string{filepath.Join("lang")})
		return NewTranslator(loader, local, fallback), nil
	})
}

func (translation *ServiceProvider) Boot(app foundation.Application) {

}
