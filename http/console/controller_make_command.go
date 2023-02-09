package console

import (
	"errors"
	"os"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
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
	}
}

// Handle Execute the console command.
func (receiver *ControllerMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Controller created successfully")

	return nil
}

func (receiver *ControllerMakeCommand) getStub() string {
	return Stubs{}.Controller()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ControllerMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyController", str.Case2Camel(name))

	return stub
}

// getPath Get the full path to the command.
func (receiver *ControllerMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/http/controllers/" + str.Camel2Case(name) + ".go"
}
