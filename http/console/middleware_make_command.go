package console

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type MiddlewareMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *MiddlewareMakeCommand) Signature() string {
	return "make:middleware"
}

// Description The console command description.
func (receiver *MiddlewareMakeCommand) Description() string {
	return "Create a new middleware class"
}

// Extend The console command extend.
func (receiver *MiddlewareMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *MiddlewareMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Middleware created successfully")

	return nil
}

func (receiver *MiddlewareMakeCommand) getStub() string {
	return Stubs{}.Middleware()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *MiddlewareMakeCommand) populateStub(stub string, name string) string {
	middlewareName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyMiddleware", str.Case2Camel(middlewareName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *MiddlewareMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	middlewareName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "http", "middleware", folderPath, str.Camel2Case(middlewareName)+".go")
}

// parseName Parse the name to get the middleware name, package name and folder path.
func (receiver *MiddlewareMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	middlewareName := segments[len(segments)-1]

	packageName := "middleware"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return middlewareName, packageName, folderPath
}
