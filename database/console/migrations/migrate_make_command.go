package migrations

import (
	"github.com/goravel/framework/database/migrations"
	"github.com/urfave/cli/v2"
	"log"
)

type MigrateMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver MigrateMakeCommand) Signature() string {
	return "make:migration"
}

//Description The console command description.
func (receiver MigrateMakeCommand) Description() string {
	return "Create a new migration file"
}

//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
func (receiver MigrateMakeCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
func (receiver MigrateMakeCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

//Handle Execute the console command.
func (receiver MigrateMakeCommand) Handle(c *cli.Context) error {
	// It's possible for the developer to specify the tables to modify in this
	// schema operation. The developer may also specify if this table needs
	// to be freshly created so we can create the appropriate migrations.
	name := c.Args().First()
	if name == "" {
		log.Fatalln(`Not enough arguments (missing: "name").`)
	}

	// We will attempt to guess the table name if this the migration has
	// "create" in the name. This will allow us to provide a convenient way
	// of creating migrations that create new tables for the application.
	table, create := TableGuesser{}.Guess(name)

	//Write the migration file to disk.
	migrations.MigrateCreator{}.Create(name, table, create)

	log.Printf("Created Migration: %s", name)

	return nil
}
