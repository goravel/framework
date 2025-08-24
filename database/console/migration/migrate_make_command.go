package migration

import (
	"fmt"
	"runtime/debug"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/str"
)

type MigrateMakeCommand struct {
	app      foundation.Application
	migrator migration.Migrator
}

func NewMigrateMakeCommand(app foundation.Application, migrator migration.Migrator) *MigrateMakeCommand {
	return &MigrateMakeCommand{app: app, migrator: migrator}
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
					return errors.MigrationNameIsRequired
				}

				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
	}

	fileName, err := r.migrator.Create(name)
	if err != nil {
		ctx.Error(errors.MigrationCreateFailed.Args(err).Error())
		return nil
	}

	ctx.Success(fmt.Sprintf("Created Migration: %s", name))

	info, _ := debug.ReadBuildInfo()
	structName := str.Of(fileName).Prepend("m_").Studly().String()
	if err = modify.GoFile(r.app.DatabasePath("kernel.go")).
		Find(match.Imports()).Modify(modify.AddImport(fmt.Sprintf("%s/database/migrations", info.Main.Path))).
		Find(match.Migrations()).Modify(modify.Register(fmt.Sprintf("&migrations.%s{}", structName))).
		Apply(); err != nil {
		ctx.Warning(errors.MigrationRegisterFailed.Args(err).Error())
		return nil
	}

	ctx.Success("Migration registered successfully")

	return nil
}
