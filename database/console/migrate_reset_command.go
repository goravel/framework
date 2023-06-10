package console

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateResetCommand struct {
	config config.Config
}

func NewMigrateResetCommand(config config.Config) *MigrateResetCommand {
	return &MigrateResetCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateResetCommand) Signature() string {
	return "migrate:reset"
}

// Description The console command description.
func (receiver *MigrateResetCommand) Description() string {
	return "Rollback all database migrations"
}

// Extend The console command extend.
func (receiver *MigrateResetCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (receiver *MigrateResetCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellowln("Please fill database config first")

		return nil
	}

	// Rollback all migrations.
	if err = m.Down(); err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration reset failed:", err.Error())

		return nil
	}

	color.Greenln("Migration reset success")

	return nil
}
