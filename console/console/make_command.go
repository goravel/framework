package console

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type MakeCommand struct {
}

func NewMakeCommand() *MakeCommand {
	return &MakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *MakeCommand) Signature() string {
	return "make:command"
}

// Description The console command description.
func (receiver *MakeCommand) Description() string {
	return "Create a new Artisan command"
}

// Extend The console command extend.
func (receiver *MakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *MakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	return file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
}

func (receiver *MakeCommand) getStub() string {
	return Stubs{}.Command()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *MakeCommand) populateStub(stub string, name string) string {
	commandName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyCommand", str.Case2Camel(commandName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *MakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	commandName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "console", "commands", folderPath, str.Camel2Case(commandName)+".go")
}

// parseName Parse the name to get the command name, package name and folder path.
func (receiver *MakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	commandName := segments[len(segments)-1]

	packageName := "commands"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return commandName, packageName, folderPath
}
