package broadcasting

import (
	"github.com/goravel/framework/broadcasting/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
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
		return NewApplication(app.MakeConfig(), app.MakeAuth(), app.MakeLog(), app.MakeQueue()), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	app.MakeQueue().Register([]queue.Job{&BroadcastJob{}})

	cfg, err := NewConfig(app.MakeConfig())
	if err != nil {
		cfg = &Config{Default: "log"}
	}

	if cfg.Auth.Enabled {
		if broadcastApp, ok := app.MakeBroadcast().(*Application); ok {
			app.MakeRoute().Post(cfg.Auth.Path, broadcastApp.Authenticate)
		}
	}

	app.MakeArtisan().Register([]contractsconsole.Command{
		console.NewChannelMakeCommand(),
	})
}
