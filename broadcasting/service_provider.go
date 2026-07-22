package broadcasting

import (
	"github.com/goravel/framework/broadcasting/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{binding.Broadcast},
		Dependencies: binding.Bindings[binding.Broadcast].Dependencies,
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Bind(binding.Broadcast, func(app foundation.Application) (any, error) {
		return NewApplication(app.MakeConfig(), app.MakeLog(), app.MakeQueue()), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	facades.Queue().Register([]queue.Job{&BroadcastJob{}})

	cfg := NewConfig(app.MakeConfig())
	auth := cfg.Auth()
	if auth.Enabled {
		broadcastInstance := app.MakeBroadcast()
		if broadcastInstance != nil {
			facades.Route().Post(auth.Path, broadcastInstance.Authenticate)
		}
	}

	facades.Artisan().Register([]contractsconsole.Command{
		console.NewChannelMakeCommand(),
	})
}
