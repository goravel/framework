package migration

import (
	"fmt"
	"path/filepath"
	"runtime/debug"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	supportconsole "github.com/goravel/framework/support/console"
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
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "The model name to be used in the migration, will create it if it doesn't exist",
			},
		},
	}
}

// Handle Executes the console command.
func (r *MigrateMakeCommand) Handle(ctx console.Context) error {
	// It's possible for the developer to specify the tables to modify in this
	// schema operation. The developer may also specify if this table needs
	// to be freshly created, so we can create the appropriate migrations.
	m, err := supportconsole.NewMake(ctx, "migration", ctx.Argument(0), filepath.Join("database", "migrations"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	migrationName := m.GetStructName()
	modelName := ctx.Option("model")

	fileName, err := r.migrator.Create(migrationName, modelName)
	if err != nil {
		ctx.Error(errors.MigrationCreateFailed.Args(err).Error())
		return nil
	}

	ctx.Success(fmt.Sprintf("Created Migration: %s", migrationName))

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
