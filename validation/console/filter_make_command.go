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

type FilterMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *FilterMakeCommand) Signature() string {
	return "make:filter"
}

// Description The console command description.
func (receiver *FilterMakeCommand) Description() string {
	return "Create a new filter class"
}

// Extend The console command extend.
func (receiver *FilterMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the filter even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *FilterMakeCommand) Handle(ctx console.Context) error {
	name, err := supportconsole.GetName(ctx, "filter", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Green().Println("Filter created successfully")

	return nil
}

func (receiver *FilterMakeCommand) getStub() string {
	return Stubs{}.Filter()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *FilterMakeCommand) populateStub(stub string, name string) string {
	ruleName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyFilter", str.Case2Camel(ruleName))
	stub = strings.ReplaceAll(stub, "DummyName", str.Camel2Case(ruleName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *FilterMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	ruleName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "filters", folderPath, str.Camel2Case(ruleName)+".go")
}

// parseName Parse the name to get the filter name, package name and folder path.
func (receiver *FilterMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	ruleName := segments[len(segments)-1]

	packageName := "filters"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return ruleName, packageName, folderPath
}
