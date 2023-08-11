package route

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.route"

type ServiceProvider struct {
}

func (route *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewRoute(app.MakeConfig()), nil
	})
}

func (route *ServiceProvider) Boot(app foundation.Application) {

}
