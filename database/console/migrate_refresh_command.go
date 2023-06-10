package console

import (
	"github.com/gookit/color"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateRefreshCommand struct {
	config config.Config
}

func NewMigrateRefreshCommand(config config.Config) *MigrateRefreshCommand {
	return &MigrateRefreshCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateRefreshCommand) Signature() string {
	return "migrate:refresh"
}

// Description The console command description.
func (receiver *MigrateRefreshCommand) Description() string {
	return "Reset and re-run all migrations"
}

// Extend The console command extend.
func (receiver *MigrateRefreshCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (receiver *MigrateRefreshCommand) Handle(ctx console.Context) error {
	migrateResetCommand := &MigrateResetCommand{}
	err := migrateResetCommand.Handle(ctx)

	if err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration reset failed:", err.Error())

		return err
	}

	migrateCommand := &MigrateCommand{}
	err = migrateCommand.Handle(ctx)

	if err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration refresh failed:", err.Error())
		return err
	}

	color.Greenln("Migration refresh success")

	return nil
}
