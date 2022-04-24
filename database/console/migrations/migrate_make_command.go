package migrations

import (
	"errors"
	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/database/migrations"
	"github.com/urfave/cli/v2"
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
func (receiver *MigrateMakeCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *MigrateMakeCommand) Handle(c *cli.Context) error {
	// It's possible for the developer to specify the tables to modify in this
	// schema operation. The developer may also specify if this table needs
	// to be freshly created, so we can create the appropriate migrations.
	name := c.Args().First()
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	// We will attempt to guess the table name if this the migration has
	// "create" in the name. This will allow us to provide a convenient way
	// of creating migrations that create new tables for the application.
	table, create := TableGuesser{}.Guess(name)

	//Write the migration file to disk.
	migrations.MigrateCreator{}.Create(name, table, create)

	color.Green.Printf("Created Migration: %s", name)

	return nil
}
