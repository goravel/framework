package migration

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
func (r *MigrateMakeCommand) Signature() string {
	return "make:migration"
}

// Description The console command description.
func (r *MigrateMakeCommand) Description() string {
	return "Create a new migration file"
}

// Extend The console command extend.
func (r *MigrateMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (r *MigrateMakeCommand) Handle(ctx console.Context) error {
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

	migrationDriver, err := GetDriver(r.config)
	if err != nil {
		return err
	}

	if err := migrationDriver.Create(name); err != nil {
		return err
	}

	color.Green().Printf("Created Migration: %s\n", name)

	return nil
}
