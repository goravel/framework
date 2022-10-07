package console

import (
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *MigrateMakeCommand) Signature() string {
	return "make:migration"
}

//Description The console command description.
func (receiver *MigrateMakeCommand) Description() string {
	return "Create a new migration file"
}

//Extend The console command extend.
func (receiver *MigrateMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *MigrateMakeCommand) Handle(ctx console.Context) error {
	// It's possible for the developer to specify the tables to modify in this
	// schema operation. The developer may also specify if this table needs
	// to be freshly created, so we can create the appropriate migrations.
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	// We will attempt to guess the table name if this the migration has
	// "create" in the name. This will allow us to provide a convenient way
	// of creating migrations that create new tables for the application.
	table, create := TableGuesser{}.Guess(name)

	//Write the migration file to disk.
	MigrateCreator{}.Create(name, table, create)

	color.Green.Printf("Created Migration: %s\n", name)

	return nil
}
