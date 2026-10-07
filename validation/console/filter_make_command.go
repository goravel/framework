package console

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"
)

type FilterMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *FilterMakeCommand) Signature() string {
	return "make:filter"
}

// Description The console command description.
func (r *FilterMakeCommand) Description() string {
	return "Create a new filter class"
}

// Extend The console command extend.
func (r *FilterMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the filter even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *FilterMakeCommand) Handle(ctx console.Context) error {
	for _, name := range supportconsole.MakeNames(ctx) {
		if err := r.makeOne(ctx, name); err != nil {
			ctx.Error(err.Error())
		}
	}

	return nil
}

func (r *FilterMakeCommand) makeOne(ctx console.Context, name string) error {
	m, err := supportconsole.NewMake(ctx, "filter", name, support.Config.Paths.Filters)
	if err != nil {
		return err
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), m.GetSignature())); err != nil {
		return err
	}

	ctx.Success("Filter created successfully")

	if env.IsBootstrapSetup() {
		err = modify.AddFilter(m.GetPackageImportPath(), fmt.Sprintf("&%s.%s{}", m.GetPackageName(), m.GetStructName()))
	} else {
		err = r.registerInKernel(m)
	}

	if err != nil {
		return errors.ValidationFilterRegisterFailed.Args(err)
	}

	ctx.Success("Filter registered successfully")

	return nil
}

func (r *FilterMakeCommand) getStub() string {
	return Stubs{}.Filter()
}

// populateStub Populate the place-holders in the command stub.
func (r *FilterMakeCommand) populateStub(stub string, packageName, structName, signature string) string {
	stub = strings.ReplaceAll(stub, "DummyFilter", structName)
	stub = strings.ReplaceAll(stub, "DummySignature", str.Of(signature).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

func (r *FilterMakeCommand) registerInKernel(m *supportconsole.Make) error {
	return modify.GoFile(path.App("providers", "validation_service_provider.go")).
		Find(match.Imports()).Modify(modify.AddImport(m.GetPackageImportPath())).
		Find(match.ValidationFilters()).Modify(modify.Register(fmt.Sprintf("&%s.%s{}", m.GetPackageName(), m.GetStructName()))).
		Apply()
}
