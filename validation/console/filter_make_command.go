package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
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
	m, err := supportconsole.NewMake(ctx, "filter", ctx.Argument(0), filepath.Join("app", "filters"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Filter created successfully")

	return nil
}

func (r *FilterMakeCommand) getStub() string {
	return Stubs{}.Filter()
}

// populateStub Populate the place-holders in the command stub.
func (r *FilterMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyFilter", structName)
	stub = strings.ReplaceAll(stub, "DummyName", str.Of(structName).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
