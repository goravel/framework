package filesystem

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.filesystem"

var application foundation.Application

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	application = app

	app.Singleton(Binding, func() (any, error) {
		return NewStorage(app.MakeConfig()), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {

}
