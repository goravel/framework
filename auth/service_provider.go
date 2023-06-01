package auth

import (
	"context"

	"github.com/goravel/framework/auth/access"
	"github.com/goravel/framework/auth/console"
	contractconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

const BindingAuth = "goravel.auth"
const BindingGate = "goravel.gate"

type ServiceProvider struct {
}

func (database *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(BindingAuth, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		return NewAuth(config.GetString("auth.defaults.guard"),
			app.MakeCache(), config, app.MakeOrm()), nil
	})
	app.Singleton(BindingGate, func(app foundation.Application) (any, error) {
		return access.NewGate(context.Background()), nil
	})
}

func (database *ServiceProvider) Boot(app foundation.Application) {
	database.registerCommands(app)
}

func (database *ServiceProvider) registerCommands(app foundation.Application) {
	app.MakeArtisan().Register([]contractconsole.Command{
		console.NewJwtSecretCommand(app.MakeConfig()),
		console.NewPolicyMakeCommand(),
	})
}
