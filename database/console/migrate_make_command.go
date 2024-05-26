package console

import (
	"errors"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
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

	// We will attempt to guess the table name if this the migration has
	// "create" in the name. This will allow us to provide a convenient way
	// of creating migrations that create new tables for the application.
	table, create := TableGuesser{}.Guess(name)

	//Write the migration file to disk.
	migrateCreator := NewMigrateCreator(receiver.config)
	if err := migrateCreator.Create(name, table, create); err != nil {
		return err
	}

	color.Green().Printf("Created Migration: %s\n", name)

	return nil
}
