package mail

import (
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/mail/console"
	"github.com/goravel/framework/support/color"
)

const Binding = "goravel.mail"

type ServiceProvider struct {
}

func (route *ServiceProvider) Register(app foundation.Application) {
	app.Bind(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleMail)
		}

		queueFacade := app.MakeQueue()
		if queueFacade == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleMail)
		}
		return NewApplication(config, queueFacade), nil
	})
}

func (route *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]consolecontract.Command{
		console.NewMailMakeCommand(),
	})

	route.registerJobs(app)
}

func (route *ServiceProvider) registerJobs(app foundation.Application) {
	queueFacade := app.MakeQueue()
	if queueFacade == nil {
		color.Warningln("Queue Facade is not initialized. Skipping job registration.")
		return
	}

	configFacade := app.MakeConfig()
	if configFacade == nil {
		color.Warningln("Config Facade is not initialized. Skipping job registration.")
		return
	}

	queueFacade.Register([]queue.Job{
		NewSendMailJob(configFacade),
	})
}
