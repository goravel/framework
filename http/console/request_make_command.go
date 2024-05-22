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

type RequestMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *RequestMakeCommand) Signature() string {
	return "make:request"
}

// Description The console command description.
func (receiver *RequestMakeCommand) Description() string {
	return "Create a new request class"
}

// Extend The console command extend.
func (receiver *RequestMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the request even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *RequestMakeCommand) Handle(ctx console.Context) error {
	name, err := supportconsole.GetName(ctx, "request", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Green().Println("Request created successfully")

	return nil
}

func (receiver *RequestMakeCommand) getStub() string {
	return Stubs{}.Request()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *RequestMakeCommand) populateStub(stub string, name string) string {
	requestName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyRequest", str.Case2Camel(requestName))
	stub = strings.ReplaceAll(stub, "DummyField", "Name string `form:\"name\" json:\"name\"`")
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *RequestMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	requestName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "http", "requests", folderPath, str.Camel2Case(requestName)+".go")
}

// parseName Parse the name to get the request name, package name and folder path.
func (receiver *RequestMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	requestName := segments[len(segments)-1]

	packageName := "requests"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return requestName, packageName, folderPath
}
