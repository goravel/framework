package console

import (
	"errors"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/urfave/cli/v2"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type EventMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *EventMakeCommand) Signature() string {
	return "make:event"
}

//Description The console command description.
func (receiver *EventMakeCommand) Description() string {
	return "Create a new event class"
}

//Extend The console command extend.
func (receiver *EventMakeCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *EventMakeCommand) Handle(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Event created successfully")

	return nil
}

func (receiver *EventMakeCommand) getStub() string {
	return EventStubs{}.Event()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *EventMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyEvent", str.Case2Camel(name))
	stub = strings.ReplaceAll(stub, "DummyName", str.Camel2Case(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *EventMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/events/" + str.Camel2Case(name) + ".go"
}
