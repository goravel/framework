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
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the policy even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *PolicyMakeCommand) Handle(ctx console.Context) error {
	name, err := supportconsole.GetName(ctx, "policy", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Green().Println("Policy created successfully")

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
