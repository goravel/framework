package database

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/database/console/migrations"
	"github.com/goravel/framework/support/facades"
)

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (database *ServiceProvider) Boot() {
	database.registerCommands()
}

//Register any application services.
func (database *ServiceProvider) Register() {
	app := Application{}
	facades.DB = app.Init()
	facades.Gorm = facades.DB
}

//registerCommands Register the given commands.
func (database *ServiceProvider) registerCommands() {
	facades.Artisan.Register([]console.Command{
		&migrations.MigrateMakeCommand{},
		&migrations.MigrateCommand{},
	})
}
