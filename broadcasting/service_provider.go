package broadcasting

import (
	"github.com/goravel/framework/broadcasting/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{binding.Broadcast},
		Dependencies: binding.Bindings[binding.Broadcast].Dependencies,
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Broadcast, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleBroadcast)
		}

		auth := app.MakeAuth()
		if auth == nil {
			return nil, errors.AuthFacadeNotSet.SetModule(errors.ModuleBroadcast)
		}

		log := app.MakeLog()
		if log == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleBroadcast)
		}

		queue := app.MakeQueue()
		if queue == nil {
			return nil, errors.QueueFacadeNotSet.SetModule(errors.ModuleBroadcast)
		}

		return NewApplication(config, auth, log, queue, app), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	queueFacade := app.MakeQueue()
	if queueFacade != nil {
		queueFacade.Register([]queue.Job{&BroadcastJob{config: app.MakeConfig(), app: app}})
	}

	cfg, err := NewConfig(app.MakeConfig())
	if err != nil {
		cfg = &Config{Default: "log"}
	}

	if cfg.Auth.Enabled {
		routeFacade := app.MakeRoute()
		if routeFacade != nil {
			if broadcastApp, ok := app.MakeBroadcast().(*Application); ok {
				routeFacade.Post(cfg.Auth.Path, broadcastApp.Authenticate)
			}
		}
	}

	artisanFacade := app.MakeArtisan()
	if artisanFacade != nil {
		artisanFacade.Register([]contractsconsole.Command{
			console.NewChannelMakeCommand(),
		})
	}
}
