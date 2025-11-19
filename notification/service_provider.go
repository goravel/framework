package notification

import (
    "github.com/goravel/framework/contracts/binding"
    "github.com/goravel/framework/contracts/foundation"
    contractsqueue "github.com/goravel/framework/contracts/queue"
    "github.com/goravel/framework/errors"
)

type ServiceProvider struct {
}

// Relationship returns the relationship of the service provider.
func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Notification,
		},
		Dependencies: binding.Bindings[binding.Notification].Dependencies,
		ProvideFor:   []string{},
	}
}

// Register registers the service provider.
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

// Boot boots the service provider, will be called after all service providers are registered.
func (r *ServiceProvider) Boot(app foundation.Application) {
    RegisterDefaultChannels(app)
    r.registerJobs(app)
}

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
