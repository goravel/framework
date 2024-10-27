package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
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
	m, err := supportconsole.NewMake(ctx, "policy", ctx.Argument(0), filepath.Join("app", "policies"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Policy created successfully")

	return nil
}

func (receiver *PolicyMakeCommand) getStub() string {
	return PolicyStubs{}.Policy()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *PolicyMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyPolicy", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
