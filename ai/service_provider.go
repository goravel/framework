package ai

import (
	"github.com/goravel/framework/ai/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.AI,
		},
		Dependencies: binding.Bindings[binding.AI].Dependencies,
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.AI, func(app foundation.Application) (any, error) {
		return NewApplication(app.MakeConfig()), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	artisan := app.MakeArtisan()
	if artisan == nil {
		return
	}
	artisan.Register([]contractsconsole.Command{
		console.NewAiDocsInstallCommand(),
		console.NewAiDocsUpdateCommand(),
	})
}
