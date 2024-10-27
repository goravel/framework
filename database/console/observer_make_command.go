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
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the observer even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ObserverMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "observer", ctx.Argument(0), filepath.Join("app", "observers"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Observer created successfully")

	return nil
}

func (receiver *ObserverMakeCommand) getStub() string {
	return Stubs{}.Observer()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ObserverMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyObserver", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
