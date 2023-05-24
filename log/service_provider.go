package log

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.log"

type ServiceProvider struct {
}

func (log *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func() (any, error) {
		return NewLogrusApplication(app.MakeConfig()), nil
	})
}

func (log *ServiceProvider) Boot(app foundation.Application) {

}
