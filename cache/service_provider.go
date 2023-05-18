package cache

import (
	"github.com/goravel/framework/cache/console"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.cache"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	config := app.MakeConfig()
	store := config.GetString("cache.default")
	app.Singleton(Binding, func() (any, error) {
		return NewApplication(config, store), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	database.registerCommands(app)
}

func (database *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]console2.Command{
		console.NewClearCommand(app.MakeCache()),
	})
}
