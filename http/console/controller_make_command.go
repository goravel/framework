package console

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

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
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "resource",
				Value: false,
				Usage: "resourceful controller",
			},
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the controller even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ControllerMakeCommand) Handle(ctx console.Context) error {
	name, err := supportconsole.GetName(ctx, "controller", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	stub := receiver.getStub()
	if ctx.OptionBool("resource") {
		stub = receiver.getResourceStub()
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(stub, name)); err != nil {
		return err
	}

	color.Green().Println("Controller created successfully")

	return nil
}

func (receiver *ControllerMakeCommand) getStub() string {
	return Stubs{}.Controller()
}

func (receiver *ControllerMakeCommand) getResourceStub() string {
	return Stubs{}.ResourceController()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *ControllerMakeCommand) populateStub(stub string, name string) string {
	controllerName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyController", str.Case2Camel(controllerName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *ControllerMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	controllerName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "http", "controllers", folderPath, str.Camel2Case(controllerName)+".go")
}

// parseName Parse the name to get the controller name, package name and folder path.
func (receiver *ControllerMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	controllerName := segments[len(segments)-1]

	packageName := "controllers"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return controllerName, packageName, folderPath
}
