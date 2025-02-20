package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type ModelMakeCommand struct {
}

func NewModelMakeCommand() *ModelMakeCommand {
	return &ModelMakeCommand{}
}

// Signature The name and signature of the console command.
func (r *ModelMakeCommand) Signature() string {
	return "make:model"
}

// Description The console command description.
func (r *ModelMakeCommand) Description() string {
	return "Create a new model class"
}

// Extend The console command extend.
func (r *ModelMakeCommand) Extend() command.Extend {
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
func (r *ModelMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "model", ctx.Argument(0), filepath.Join("app", "models"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Model created successfully")

	return nil
}

func (r *ModelMakeCommand) getStub() string {
	return Stubs{}.Model()
}

// populateStub Populate the place-holders in the command stub.
func (r *ModelMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyModel", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
