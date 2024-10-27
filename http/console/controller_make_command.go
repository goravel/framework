package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type ControllerMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *ControllerMakeCommand) Signature() string {
	return "make:controller"
}

// Description The console command description.
func (receiver *ControllerMakeCommand) Description() string {
	return "Create a new controller class"
}

// Extend The console command extend.
func (receiver *ControllerMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "resource",
				Value: false,
				Usage: "resourceful controller",
			},
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the controller even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ControllerMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "controller", ctx.Argument(0), filepath.Join("app", "http", "controllers"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	stub := receiver.getStub()
	if ctx.OptionBool("resource") {
		stub = receiver.getResourceStub()
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(stub, m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Controller created successfully")

	return nil
}

func (receiver *ControllerMakeCommand) getStub() string {
	return Stubs{}.Controller()
}

func (receiver *ControllerMakeCommand) getResourceStub() string {
	return Stubs{}.ResourceController()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ControllerMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyController", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
