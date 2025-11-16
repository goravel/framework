package migration

import (
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/env"
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
	make, err := supportconsole.NewMake(ctx, "command", ctx.Argument(0), support.Config.Paths.Migration)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	fileName, err := r.migrator.Create(make.GetName())
	if err != nil {
		ctx.Error(errors.MigrationCreateFailed.Args(err).Error())
		return nil
	}

	ctx.Success(fmt.Sprintf("Created Migration: %s", make.GetName()))

	structName := str.Of(fileName).Prepend("m_").Studly().String()
	if env.IsBootstrapSetup() {
		err = modify.AddMigration(make.GetPackageImportPath(), fmt.Sprintf("&%s.%s{}", make.GetPackageName(), structName))
	} else {
		err = r.registerInKernel(make.GetPackageImportPath(), structName)
	}

	if err != nil {
		ctx.Error(errors.MigrationRegisterFailed.Args(err).Error())
		return nil
	}

	ctx.Success("Migration registered successfully")

	return nil
}

func (r *MigrateMakeCommand) registerInKernel(pkg, structName string) error {
	return modify.GoFile(r.app.DatabasePath("kernel.go")).
		Find(match.Imports()).Modify(modify.AddImport(pkg)).
		Find(match.Migrations()).Modify(modify.Register(fmt.Sprintf("&migrations.%s{}", structName))).
		Apply()
}
