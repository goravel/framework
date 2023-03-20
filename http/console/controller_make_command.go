package console

import (
	"errors"
	"os"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

// Default package name if a sub-folder is not provided
const defaultPackageName string = "controllers"

type ControllerMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *ControllerMakeCommand) Signature() string {
	return "make:controller"
}

// Description The console command description.
func (receiver *ControllerMakeCommand) Description() string {
	return "Create a new controller class"
}

// Extend The console command extend.
func (receiver *ControllerMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *ControllerMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Controller created successfully")

	return nil
}

func (receiver *ControllerMakeCommand) getStub() string {
	return Stubs{}.Controller()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ControllerMakeCommand) populateStub(stub string, name string) string {
	controllerName, packageName, _ := parseName(name)

	stub = strings.ReplaceAll(stub, "DummyController", str.Case2Camel(controllerName))
	stub = strings.ReplaceAll(stub, "dummy_package", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *ControllerMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	controllerName, _, folderPath := parseName(name)

	if folderPath != "" {
		folderPath = folderPath + "/"
	}

	return pwd + "/app/http/controllers/" + folderPath + str.Camel2Case(controllerName) + ".go"
}

func parseName(name string) (string, string, string) {

	parts := strings.Split(name, "/")

	controllerName := parts[len(parts)-1]
	packageName := defaultPackageName
	filePath := strings.ToLower(strings.Join(parts[:len(parts)-1], "/"))

	if len(parts) > 1 {
		packageName = strings.ToLower(parts[len(parts)-2])
	}

	return controllerName, packageName, filePath
}
