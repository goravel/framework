package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type ListenerMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *ListenerMakeCommand) Signature() string {
	return "make:listener"
}

// Description The console command description.
func (receiver *ListenerMakeCommand) Description() string {
	return "Create a new listener class"
}

// Extend The console command extend.
func (receiver *ListenerMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the listener even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ListenerMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "listener", ctx.Argument(0), filepath.Join("app", "listeners"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Listener created successfully")

	return nil
}

func (receiver *ListenerMakeCommand) getStub() string {
	return Stubs{}.Listener()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ListenerMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyListener", structName)
	stub = strings.ReplaceAll(stub, "DummyName", str.Of(structName).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
