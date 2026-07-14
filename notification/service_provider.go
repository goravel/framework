package notification

import (
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/notification/console"
	"github.com/goravel/framework/support/color"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Notification,
		},
		Dependencies: binding.Bindings[binding.Notification].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Bind(binding.Notification, func(app foundation.Application) (any, error) {
		logger := app.MakeLog()
		if logger == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleNotification)
		}

		mail := app.MakeMail()
		if mail == nil {
			return nil, errors.MailFacadeNotSet.SetModule(errors.ModuleNotification)
		}

		orm := app.MakeOrm()
		if orm == nil {
			return nil, errors.OrmFacadeNotSet.SetModule(errors.ModuleNotification)
		}

		q := app.MakeQueue()
		if q == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleNotification)
		}

		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleNotification)
		}

		manager := NewManager(logger, q)
		manager.Extend(channels.NewMailChannel(mail, logger))
		manager.Extend(channels.NewDatabaseChannel(orm, logger, config))

		return manager, nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]contractsconsole.Command{
		console.NewNotificationMakeCommand(),
		console.NewNotificationTableCommand(),
	})

	r.registerJobs(app)
}

func (r *ServiceProvider) registerJobs(app foundation.Application) {
	queueFacade := app.MakeQueue()
	if queueFacade == nil {
		color.Warningln("Queue Facade is not initialized. Skipping job registration.")
		return
	}

	notificationFacade := app.MakeNotification()
	if notificationFacade == nil {
		color.Warningln("Notification Facade is not initialized. Skipping job registration.")
		return
	}

	manager, ok := notificationFacade.(*Manager)
	if !ok {
		color.Warningln("Notification Facade is not a *Manager. Skipping job registration.")
		return
	}

	queueFacade.Register([]contractsqueue.Job{
		NewDispatchJob(manager),
	})
}
