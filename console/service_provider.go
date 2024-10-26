package console

import (
	"github.com/goravel/framework/console/console"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/color"
)

const Binding = "goravel.console"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		name := "Goravel Framework"
		usage := app.Version()
		usageText := "artisan [global options] command [options] [arguments...]"
		return NewApplication(name, usage, usageText, app.Version(), true), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	receiver.registerCommands(app)
}

func (receiver *ServiceProvider) registerCommands(app foundation.Application) {
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
