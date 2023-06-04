package cache

import (
	"github.com/goravel/framework/cache/console"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.cache"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		store := config.GetString("cache.default")

		return NewApplication(config, store), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	database.registerCommands(app)
}

func (database *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]contractsconsole.Command{
		console.NewClearCommand(app.MakeCache()),
	})
}
