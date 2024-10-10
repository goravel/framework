package queue

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	queueConsole "github.com/goravel/framework/queue/console"
)

const Binding = "goravel.queue"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleQueue)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleQueue)
		}

		return NewApplication(config, log), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]console.Command{
		&queueConsole.JobMakeCommand{},
	})
}
