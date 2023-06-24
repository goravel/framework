package translation

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.translation"

type ServiceProvider struct {
}

func (trans *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewTranslation(app.MakeConfig()), nil
	})
}

func (trans *ServiceProvider) Boot(app foundation.Application) {

}
