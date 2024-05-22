package console

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type EventMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *EventMakeCommand) Signature() string {
	return "make:event"
}

// Description The console command description.
func (receiver *EventMakeCommand) Description() string {
	return "Create a new event class"
}

// Extend The console command extend.
func (receiver *EventMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the event even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *EventMakeCommand) Handle(ctx console.Context) error {
	name, err := supportconsole.GetName(ctx, "event", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Green().Println("Event created successfully")

	return nil
}

func (receiver *EventMakeCommand) getStub() string {
	return EventStubs{}.Event()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *EventMakeCommand) populateStub(stub string, name string) string {
	eventName, packageName, _ := receiver.parseName(name)
	stub = strings.ReplaceAll(stub, "DummyEvent", str.Case2Camel(eventName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *EventMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	eventName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "events", folderPath, str.Camel2Case(eventName)+".go")
}

// parseName Parse the name to get the event name, package name and folder path.
func (receiver *EventMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	eventName := segments[len(segments)-1]

	packageName := "events"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return eventName, packageName, folderPath
}
