package migration

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/support/color"
)

type MigrateCommand struct {
	driver migration.Driver
}

func NewMigrateCommand(config config.Config, schema migration.Schema) *MigrateCommand {
	driver, err := GetDriver(config, schema)
	if err != nil {
		color.Red().Println(err.Error())
		return nil
	}

	return &MigrateCommand{
		driver: driver,
	}
}

// Signature The name and signature of the console command.
func (r *MigrateCommand) Signature() string {
	return "migrate"
}

// Description The console command description.
func (r *MigrateCommand) Description() string {
	return "Run the database migrations"
}

// Extend The console command extend.
func (r *MigrateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (r *MigrateCommand) Handle(ctx console.Context) error {
	if err := r.driver.Run(); err != nil {
		color.Red().Println("Migration failed:", err.Error())
		return nil
	}

	color.Green().Println("Migration success")

	return nil
}
