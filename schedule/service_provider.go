package schedule

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	scheduleconsole "github.com/goravel/framework/schedule/console"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingSchedule, func(app foundation.Application) (any, error) {
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

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]console.Command{
		scheduleconsole.NewList(app.MakeSchedule()),
		scheduleconsole.NewRun(app.MakeSchedule()),
	})
}
