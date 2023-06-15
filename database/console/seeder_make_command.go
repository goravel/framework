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

type SeederMakeCommand struct {
}

func NewSeederMakeCommand() *SeederMakeCommand {
	return &SeederMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *SeederMakeCommand) Signature() string {
	return "make:seeder"
}

// Description The console command description.
func (receiver *SeederMakeCommand) Description() string {
	return "Create a new seeder class"
}

// Extend The console command extend.
func (receiver *SeederMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *SeederMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Seeder created successfully")

	return nil
}

func (receiver *SeederMakeCommand) getStub() string {
	return Stubs{}.Seeder()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *SeederMakeCommand) populateStub(stub string, name string) string {
	modelName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummySeeder", str.Case2Camel(modelName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *SeederMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	modelName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "database", "seeders", folderPath, str.Camel2Case(modelName)+".go")
}

// parseName Parse the name to get the model name, package name and folder path.
func (receiver *SeederMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	modelName := segments[len(segments)-1]

	packageName := "seeders"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return modelName, packageName, folderPath
}
