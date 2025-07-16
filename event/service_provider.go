package event

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	eventConsole "github.com/goravel/framework/event/console"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Event,
		},
		Dependencies: []string{
			binding.Queue,
		},
		ProvideFor: []string{},
	}
}

var App foundation.Application

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Event, func(app foundation.Application) (any, error) {
		App = app
		queueFacade := app.MakeQueue()
		if queueFacade == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleEvent)
		}

		return NewApplication(queueFacade), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]console.Command{
		&eventConsole.EventMakeCommand{},
		&eventConsole.ListenerMakeCommand{},
	})
}
