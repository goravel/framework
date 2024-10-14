package schedule

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

const Binding = "goravel.schedule"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleSchedule)
		}

		artisan := app.MakeArtisan()
		if artisan == nil {
			return nil, errors.ArtisanFacadeNotSet.SetModule(errors.ModuleSchedule)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleSchedule)
		}

		cache := app.MakeCache()
		if cache == nil {
			return nil, errors.CacheFacadeNotSet.SetModule(errors.ModuleSchedule)
		}

		return NewApplication(artisan, cache, log, config.GetBool("app.debug")), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {

}
