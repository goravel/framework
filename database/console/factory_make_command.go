package console

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
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
	}
}

// Handle Execute the console command.
func (receiver *FactoryMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Factory created successfully")

	return nil
}

func (receiver *FactoryMakeCommand) getStub() string {
	return Stubs{}.Factory()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *FactoryMakeCommand) populateStub(stub string, name string) string {
	modelName, packageName, _ := parseName(name, "factories")

	stub = strings.ReplaceAll(stub, "DummyFactory", str.Case2Camel(modelName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *FactoryMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	modelName, _, folderPath := parseName(name, "factories")

	return filepath.Join(pwd, "database", "factories", folderPath, str.Camel2Case(modelName)+".go")
}

// parseName Parse the name to get the model name, package name and folder path.
func parseName(name string, packageName string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	modelName := segments[len(segments)-1]

	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return modelName, packageName, folderPath
}
