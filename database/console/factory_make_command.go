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

type FactoryMakeCommand struct {
}

func NewFactoryMakeCommand() *FactoryMakeCommand {
	return &FactoryMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *FactoryMakeCommand) Signature() string {
	return "make:factory"
}

// Description The console command description.
func (receiver *FactoryMakeCommand) Description() string {
	return "Create a new factory class"
}

// Extend The console command extend.
func (receiver *FactoryMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the factory even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *FactoryMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "factory", ctx.Argument(0), filepath.Join("database", "factories"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Factory created successfully")

	return nil
}

func (receiver *FactoryMakeCommand) getStub() string {
	return Stubs{}.Factory()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *FactoryMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyFactory", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
