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

type ListenerMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *ListenerMakeCommand) Signature() string {
	return "make:listener"
}

//Description The console command description.
func (receiver *ListenerMakeCommand) Description() string {
	return "Create a new listener class"
}

//Extend The console command extend.
func (receiver *ListenerMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *ListenerMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Listener created successfully")

	return nil
}

func (receiver *ListenerMakeCommand) getStub() string {
	return ListenerStubs{}.Listener()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *ListenerMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyListener", str.Case2Camel(name))
	stub = strings.ReplaceAll(stub, "DummyName", str.Camel2Case(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *ListenerMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/listeners/" + str.Camel2Case(name) + ".go"
}
