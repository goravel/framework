package hash

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.hash"

type ServiceProvider struct {
}

func (hash *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func() (any, error) {
		return NewApplication(), nil
	})
}

func (hash *ServiceProvider) Boot(app foundation.Application) {

}
