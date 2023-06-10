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

type MigrateFreshCommand struct {
	config config.Config
}

func NewMigrateFreshCommand(config config.Config) *MigrateFreshCommand {
	return &MigrateFreshCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateFreshCommand) Signature() string {
	return "migrate:fresh"
}

// Description The console command description.
func (receiver *MigrateFreshCommand) Description() string {
	return "Drop all tables and re-run all migrations"
}

// Extend The console command extend.
func (receiver *MigrateFreshCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (receiver *MigrateFreshCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellowln("Please fill database config first")
		return nil
	}

	if err = m.Drop(); err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration failed:", err.Error())
		return nil
	}

	m2, err2 := getMigrate(receiver.config)
	if err2 != nil {
		return err2
	}
	if m2 == nil {
		color.Yellowln("Please fill database config first")
		return nil
	}

	if err2 = m2.Up(); err2 != nil && err2 != migrate.ErrNoChange {
		color.Redln("Migration failed:", err2.Error())
		return nil
	}

	color.Greenln("Migration fresh success")

	return nil
}
