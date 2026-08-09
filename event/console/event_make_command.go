package console

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type EventMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *EventMakeCommand) Signature() string {
	return "make:event"
}

// Description The console command description.
func (r *EventMakeCommand) Description() string {
	return "Create a new event class"
}

// Extend The console command extend.
func (r *EventMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the event even if it already exists",
			},
			&command.BoolFlag{
				Name:  "broadcast",
				Usage: "Generate an event that implements broadcasting.ShouldBroadcast",
			},
			&command.BoolFlag{
				Name:  "now",
				Usage: "Generate an event that also implements broadcasting.ShouldBroadcastNow",
			},
		},
	}
}

// Handle Execute the console command.
func (r *EventMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "event", ctx.Argument(0), support.Config.Paths.Events)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(ctx), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Event created successfully")

	return nil
}

func (r *EventMakeCommand) getStub(ctx console.Context) string {
	if ctx.OptionBool("broadcast") {
		if ctx.OptionBool("now") {
			return Stubs{}.EventBroadcastNow()
		}
		return Stubs{}.EventBroadcast()
	}

	return Stubs{}.Event()
}

// populateStub Populate the place-holders in the command stub.
func (r *EventMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyEvent", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
