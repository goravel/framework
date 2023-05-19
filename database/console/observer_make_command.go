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

type ObserverMakeCommand struct {
}

func NewObserverMakeCommand() *ObserverMakeCommand {
	return &ObserverMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *ObserverMakeCommand) Signature() string {
	return "make:observer"
}

// Description The console command description.
func (receiver *ObserverMakeCommand) Description() string {
	return "Create a new observer class"
}

// Extend The console command extend.
func (receiver *ObserverMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *ObserverMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Observer created successfully")

	return nil
}

func (receiver *ObserverMakeCommand) getStub() string {
	return Stubs{}.Observer()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ObserverMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyObserver", str.Case2Camel(name))

	return stub
}

// getPath Get the full path to the command.
func (receiver *ObserverMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/observers/" + str.Camel2Case(name) + ".go"
}
