package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type FilterMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *FilterMakeCommand) Signature() string {
	return "make:filter"
}

// Description The console command description.
func (receiver *FilterMakeCommand) Description() string {
	return "Create a new filter class"
}

// Extend The console command extend.
func (receiver *FilterMakeCommand) Extend() command.Extend {
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
func (receiver *FilterMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "filter", ctx.Argument(0), filepath.Join("app", "filters"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Filter created successfully")

	return nil
}

func (receiver *FilterMakeCommand) getStub() string {
	return Stubs{}.Filter()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *FilterMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyFilter", structName)
	stub = strings.ReplaceAll(stub, "DummyName", str.Of(structName).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
