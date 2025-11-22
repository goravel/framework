package notification

import (
    "github.com/goravel/framework/contracts/binding"
    "github.com/goravel/framework/contracts/foundation"
    contractsqueue "github.com/goravel/framework/contracts/queue"
    "github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

// Relationship declares bindings and dependencies for the notification service provider.
func (r *ServiceProvider) Relationship() binding.Relationship {
    return binding.Relationship{
        Bindings: []string{
            binding.Notification,
        },
        Dependencies: binding.Bindings[binding.Notification].Dependencies,
        ProvideFor:   []string{},
    }
}

// Register binds the notification Application into the container using required facades.
func (r *ServiceProvider) Register(app foundation.Application) {
    app.Bind(binding.Notification, func(app foundation.Application) (any, error) {
        config := app.MakeConfig()
        if config == nil {
            return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleMail)
        }

		queue := app.MakeQueue()
		if queue == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleQueue)
		}

		mail := app.MakeMail()
		if mail == nil {
			return nil, errors.New("Mail facade not set")
		}

		db := app.MakeDB()
		if db == nil {
			return nil, errors.DBFacadeNotSet.SetModule(errors.ModuleDB)
		}

        return NewApplication(config, queue, db, mail)
    })
}

// Boot initializes built-in channels and registers jobs once all providers are registered.
func (r *ServiceProvider) Boot(app foundation.Application) {
    RegisterDefaultChannels()
    r.registerJobs(app)
}

// registerJobs registers the SendNotificationJob with the queue facade if available.
func (r *ServiceProvider) registerJobs(app foundation.Application) {
    queueFacade := app.MakeQueue()
    if queueFacade == nil {
        return
    }

    configFacade := app.MakeConfig()
    if configFacade == nil {
        return
    }

    mailFacade := app.MakeMail()
    if mailFacade == nil {
        return
    }

    dbFacade := app.MakeDB()
    if dbFacade == nil {
        return
    }

    queueFacade.Register([]contractsqueue.Job{
        NewSendNotificationJob(configFacade, dbFacade, mailFacade),
    })
}
