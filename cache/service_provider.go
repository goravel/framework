package cache

import (
	"github.com/goravel/framework/cache/console"
	"github.com/goravel/framework/console/support"
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (database *ServiceProvider) Boot() {
	database.registerCommands()
}

//Register Register any application services.
func (database *ServiceProvider) Register() {
	app := Application{}
	facades.Cache = app.Init()
}

//registerCommands Register the given commands.
func (database *ServiceProvider) registerCommands() {
	facades.Artisan.Register([]support.Command{
		console.ClearCommand{},
	})
}
