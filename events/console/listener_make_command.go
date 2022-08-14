package console

import (
	"errors"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/events/console/stubs"
	"github.com/goravel/framework/support"
	"github.com/urfave/cli/v2"
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
func (receiver *ListenerMakeCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *ListenerMakeCommand) Handle(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	support.Helpers{}.CreateFile(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Listener created successfully")

	return nil
}

func (receiver *ListenerMakeCommand) getStub() string {
	return stubs.ListenerStubs{}.Listener()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *ListenerMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyListener", support.Helpers{}.Case2Camel(name))
	stub = strings.ReplaceAll(stub, "DummyName", support.Helpers{}.Camel2Case(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *ListenerMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/listeners/" + support.Helpers{}.Camel2Case(name) + ".go"
}
