package console

import (
	"github.com/gookit/color"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateFreshCommand struct {
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

	m, err := getMigrate()
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellowln("Please fill database config first")
		return nil
	}

	// Drop all tables
	if err = m.Drop(); err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration failed:", err.Error())
		return err
	}

	// Run all migrations
	if err = m.Up(); err != nil && err != migrate.ErrNoChange  {
		color.Redln("Migration failed:", err.Error())
		return err
	}

	color.Greenln("Migration fresh success")

	return nil
}
