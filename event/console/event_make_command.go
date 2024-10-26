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
	m, err := supportconsole.NewMake(ctx, "event", ctx.Argument(0), filepath.Join("app", "events"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Event created successfully")

	return nil
}

func (receiver *EventMakeCommand) getStub() string {
	return Stubs{}.Event()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *EventMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyEvent", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
