package mail

import (
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/mail/console"
	"github.com/goravel/framework/support/color"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Mail,
		},
		Dependencies: binding.Bindings[binding.Mail].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	// The Mail facade resolves a new Application on every call, so the fake
	// state is shared by all of them: a Fake() call affects the next ones,
	// including the mailer of the notification mail channel.
	fake := NewFake()

	app.Bind(binding.Mail, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleMail)
		}

		queue := app.MakeQueue()
		if queue == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleMail)
		}

		application, err := NewApplication(config, queue)
		if err != nil {
			return nil, err
		}
		application.fake = fake

		return application, nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		console.NewMailMakeCommand(),
	})

	r.registerJobs(app)
}

func (r *ServiceProvider) registerJobs(app foundation.Application) {
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

	queueFacade.Register([]contractsqueue.Job{
		NewSendMailJob(configFacade),
	})
}
