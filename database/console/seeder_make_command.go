package console

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type SeederMakeCommand struct {
	app foundation.Application
}

func NewSeederMakeCommand(app foundation.Application) *SeederMakeCommand {
	return &SeederMakeCommand{
		app: app,
	}
}

// Signature The name and signature of the console command.
func (r *SeederMakeCommand) Signature() string {
	return "make:seeder"
}

// Description The console command description.
func (r *SeederMakeCommand) Description() string {
	return "Create a new seeder class"
}

// Extend The console command extend.
func (r *SeederMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the seeder even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *SeederMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "seeder", ctx.Argument(0), support.Config.Paths.Seeder)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err = file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), m.GetSignature())); err != nil {
		return err
	}

	ctx.Success("Seeder created successfully")

	if err = modify.GoFile(r.app.DatabasePath("kernel.go")).
		Find(match.Imports()).Modify(modify.AddImport(m.GetPackageImportPath())).
		Find(match.Seeders()).Modify(modify.Register(fmt.Sprintf("&%s.%s{}", m.GetPackageName(), m.GetStructName()))).
		Apply(); err != nil {
		ctx.Warning(errors.DatabaseSeederRegisterFailed.Args(err).Error())
		return nil
	}

	ctx.Success("Seeder registered successfully")

	return nil
}

func (r *SeederMakeCommand) getStub() string {
	return Stubs{}.Seeder()
}

// populateStub Populate the place-holders in the command stub.
func (r *SeederMakeCommand) populateStub(stub string, packageName, structName, signature string) string {
	stub = strings.ReplaceAll(stub, "DummySeeder", structName)
	stub = strings.ReplaceAll(stub, "DummySignature", signature)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
