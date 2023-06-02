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
func (receiver *EventMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *EventMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Event created successfully")

	return nil
}

func (receiver *EventMakeCommand) getStub() string {
	return EventStubs{}.Event()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *EventMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyEvent", str.Case2Camel(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *EventMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/events/" + str.Camel2Case(name) + ".go"
}
