package event

import (
	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	eventConsole "github.com/goravel/framework/event/console"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingEvent, func(app foundation.Application) (any, error) {
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
