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

func (event *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(contracts.BindingEvent, func(app foundation.Application) (any, error) {
		queueFacade := app.MakeQueue()
		if queueFacade == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleEvent)
		}

		return NewApplication(queueFacade), nil
	})
}

func (event *ServiceProvider) Boot(app foundation.Application) {
	event.registerCommands(app)
}

func (event *ServiceProvider) registerCommands(app foundation.Application) {
	app.Commands([]console.Command{
		&eventConsole.EventMakeCommand{},
		&eventConsole.ListenerMakeCommand{},
	})
}
