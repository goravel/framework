package testing

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.testing"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewApplication(app), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {

}
