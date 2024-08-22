package console

import (
	"errors"
	"fmt"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration"
	"github.com/goravel/framework/support/color"
)

type MigrateMakeCommand struct {
	config config.Config
}

func NewMigrateMakeCommand(config config.Config) *MigrateMakeCommand {
	return &MigrateMakeCommand{config: config}
}

// Signature The name and signature of the console command.
func (receiver *MigrateMakeCommand) Signature() string {
	return "make:migration"
}

// Description The console command description.
func (receiver *MigrateMakeCommand) Description() string {
	return "Create a new migration file"
}

// Extend The console command extend.
func (receiver *MigrateMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *MigrateMakeCommand) Handle(ctx console.Context) error {
	// It's possible for the developer to specify the tables to modify in this
	// schema operation. The developer may also specify if this table needs
	// to be freshly created, so we can create the appropriate migrations.
	name := ctx.Argument(0)
	if name == "" {
		var err error
		name, err = ctx.Ask("Enter the migration name", console.AskOption{
			Validate: func(s string) error {
				if s == "" {
					return errors.New("the migration name cannot be empty")
				}

				return nil
			},
		})
		if err != nil {
			return err
		}
	}

	var migrationDriver contractsmigration.Driver
	driver := receiver.config.GetString("database.migration.driver")

	switch driver {
	case contractsmigration.DriverDefault:
		migrationDriver = migration.NewDefaultDriver()
	case contractsmigration.DriverSql:
		migrationDriver = migration.NewSqlDriver(receiver.config)
	default:
		return fmt.Errorf("unsupported migration driver: %s", driver)
	}

	// Write the migration file to disk.
	if err := migrationDriver.Create(name); err != nil {
		return err
	}

	color.Green().Printf("Created Migration: %s\n", name)

	return nil
}
