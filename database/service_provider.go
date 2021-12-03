package database

import (
	"github.com/goravel/framework/console/support"
	"github.com/goravel/framework/database/console/migrations"
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (database *ServiceProvider) Boot() {

}

//Register Register any application services.
func (database *ServiceProvider) Register() {
	app := Application{}
	facades.DB = app.Init()

	database.registerCommands()
}

//registerCommands Register the given commands.
func (database *ServiceProvider) registerCommands() {
	facades.Artisan.Register([]support.Command{
		migrations.MigrateMakeCommand{},
		migrations.MigrateCommand{},
	})
}
