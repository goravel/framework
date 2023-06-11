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

type PolicyMakeCommand struct {
}

func NewPolicyMakeCommand() *PolicyMakeCommand {
	return &PolicyMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *PolicyMakeCommand) Signature() string {
	return "make:policy"
}

// Description The console command description.
func (receiver *PolicyMakeCommand) Description() string {
	return "Create a new policy class"
}

// Extend The console command extend.
func (receiver *PolicyMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *PolicyMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Policy created successfully")

	return nil
}

func (receiver *PolicyMakeCommand) getStub() string {
	return PolicyStubs{}.Policy()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *PolicyMakeCommand) populateStub(stub string, name string) string {
	policyName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyPolicy", str.Case2Camel(policyName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *PolicyMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	policyName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "policies", folderPath, str.Camel2Case(policyName)+".go")
}

// parseName Parse the name to get the policy name, package name and folder path.
func (receiver *PolicyMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	policyName := segments[len(segments)-1]

	packageName := "policies"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return policyName, packageName, folderPath
}
