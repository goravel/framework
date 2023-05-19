package database

import (
	"context"
	"fmt"

	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/console"
)

const Binding = "goravel.orm"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func() (any, error) {
		config := app.MakeConfig()
		defaultConnection := config.GetString("database.default")

		orm, err := InitializeOrm(context.Background(), config, defaultConnection)
		if err != nil {
			return nil, fmt.Errorf("[Orm] Init %s connection error: %v", defaultConnection, err)
		}

		return orm, nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	database.registerCommands(app)
}

func (database *ServiceProvider) registerCommands(app foundation.Application) {
	config := app.MakeConfig()
	app.MakeArtisan().Register([]consolecontract.Command{
		console.NewMigrateMakeCommand(config),
		console.NewMigrateCommand(config),
		console.NewMigrateRollbackCommand(config),
		console.NewModelMakeCommand(),
		console.NewObserverMakeCommand(),
	})
}
