package console

import (
	"os"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type FactoryMakeCommand struct {
}

func NewFactoryMakeCommand() *FactoryMakeCommand {
	return &FactoryMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *FactoryMakeCommand) Signature() string {
	return "make:factory"
}

// Description The console command description.
func (receiver *FactoryMakeCommand) Description() string {
	return "Create a new model factory"
}

// Extend The console command extend.
func (receiver *FactoryMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *FactoryMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Model created successfully")

	return nil
}

func (receiver *FactoryMakeCommand) getStub() string {
	return Stubs{}.Model()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *FactoryMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyModel", str.Case2Camel(name))

	return stub
}

// getPath Get the full path to the command.
func (receiver *FactoryMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/models/" + str.Camel2Case(name) + ".go"
}
