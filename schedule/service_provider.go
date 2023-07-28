package schedule

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.schedule"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewApplication(app.MakeArtisan(), app.MakeCache(), app.MakeLog()), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {

}
