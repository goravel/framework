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

type ModelMakeCommand struct {
}

func NewModelMakeCommand() *ModelMakeCommand {
	return &ModelMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *ModelMakeCommand) Signature() string {
	return "make:model"
}

// Description The console command description.
func (receiver *ModelMakeCommand) Description() string {
	return "Create a new model class"
}

// Extend The console command extend.
func (receiver *ModelMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the model even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ModelMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "model", ctx.Argument(0), filepath.Join("app", "models"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Model created successfully")

	return nil
}

func (receiver *ModelMakeCommand) getStub() string {
	return Stubs{}.Model()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ModelMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyModel", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
