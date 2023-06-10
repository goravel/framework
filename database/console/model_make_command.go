package console

import (
	"os"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
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
	}
}

// Handle Execute the console command.
func (receiver *ModelMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Model created successfully")

	return nil
}

func (receiver *ModelMakeCommand) getStub() string {
	return Stubs{}.Model()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ModelMakeCommand) populateStub(stub string, name string) string {
	modelName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyModel", str.Case2Camel(modelName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *ModelMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	modelName, _, folderPath := receiver.parseName(name)

	if folderPath != "" {
		folderPath = folderPath + "/"
	}

	return pwd + "/app/models/" + folderPath + str.Case2Camel(modelName) + ".go"
}

// parseName Parse the name to get the model name, package name and folder path.
func (receiver *ModelMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	modelName := segments[len(segments)-1]

	packageName := "models"

	if len(segments) > 1 {
		packageName = strings.Join(segments[:len(segments)-1], "/")
	}

	folderPath := ""

	if packageName != "models" {
		folderPath = packageName
	}

	return modelName, packageName, folderPath
}
