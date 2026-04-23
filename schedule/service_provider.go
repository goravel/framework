package schedule

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	scheduleconsole "github.com/goravel/framework/schedule/console"
	"github.com/goravel/framework/support/color"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Schedule,
		},
		Dependencies: binding.Bindings[binding.Schedule].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Schedule, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleSchedule)
		}

		artisan := app.MakeArtisan()
		if artisan == nil {
			return nil, errors.ConsoleFacadeNotSet.SetModule(errors.ModuleSchedule)
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
	schedule := app.MakeSchedule()
	if schedule == nil {
		color.Warningln(errors.ScheduleFacadeNotSet.Error())
		return
	}

	app.Commands([]console.Command{
		scheduleconsole.NewList(schedule),
		scheduleconsole.NewRun(schedule),
	})
}

func (r *ServiceProvider) Runners(app foundation.Application) []foundation.Runner {
	if app.MakeArtisan() == nil {
		return nil
	}

	schedule := app.MakeSchedule()
	if schedule == nil {
		return nil
	}

	return []foundation.Runner{NewScheduleRunner(app.MakeConfig(), schedule)}
}
