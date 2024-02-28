package session

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.session"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewManager(app), nil
	})
}

func (receiver *ServiceProvider) Boot(foundation.Application) {
}
