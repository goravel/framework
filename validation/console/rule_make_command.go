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

type RuleMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *RuleMakeCommand) Signature() string {
	return "make:rule"
}

// Description The console command description.
func (receiver *RuleMakeCommand) Description() string {
	return "Create a new rule class"
}

// Extend The console command extend.
func (receiver *RuleMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the rule even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *RuleMakeCommand) Handle(ctx console.Context) error {
	name, err := supportconsole.GetName(ctx, "rule", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Green().Println("Rule created successfully")

	return nil
}

func (receiver *RuleMakeCommand) getStub() string {
	return Stubs{}.Request()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *RuleMakeCommand) populateStub(stub string, name string) string {
	ruleName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyRule", str.Case2Camel(ruleName))
	stub = strings.ReplaceAll(stub, "DummyName", str.Camel2Case(ruleName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *RuleMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	ruleName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "rules", folderPath, str.Camel2Case(ruleName)+".go")
}

// parseName Parse the name to get the rule name, package name and folder path.
func (receiver *RuleMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	ruleName := segments[len(segments)-1]

	packageName := "rules"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return ruleName, packageName, folderPath
}
