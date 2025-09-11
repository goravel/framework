package console

import (
	"github.com/goravel/framework/console/console"
	"github.com/goravel/framework/contracts/binding"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/color"
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Artisan,
		},
		Dependencies: binding.Bindings[binding.Artisan].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Artisan, func(app foundation.Application) (any, error) {
		name := "artisan"
		usage := "Goravel Framework"
		usageText := "artisan [global options] command [options] [arguments...]"

		return NewApplication(name, usage, usageText, app.Version(), true), nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	r.registerCommands(app)
}

func (r *ServiceProvider) registerCommands(app foundation.Application) {
	artisanFacade := app.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln("Artisan Facade is not initialized. Skipping command registration.")
		return
	}

	configFacade := app.MakeConfig()
	if configFacade == nil {
		color.Warningln("Config Facade is not initialized. Skipping certain command registrations.")
		return
	}

	artisanFacade.Register([]consolecontract.Command{
		console.NewListCommand(artisanFacade),
		console.NewKeyGenerateCommand(configFacade),
		console.NewMakeCommand(),
		console.NewBuildCommand(configFacade),
	})
}
