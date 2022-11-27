package console

import (
	"errors"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"

	"github.com/gookit/color"
)

type PolicyMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *PolicyMakeCommand) Signature() string {
	return "make:policy"
}

//Description The console command description.
func (receiver *PolicyMakeCommand) Description() string {
	return "Create a new policy class"
}

//Extend The console command extend.
func (receiver *PolicyMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *PolicyMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Policy created successfully")

	return nil
}

func (receiver *PolicyMakeCommand) getStub() string {
	return PolicyStubs{}.Policy()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *PolicyMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyPolicy", str.Case2Camel(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *PolicyMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/policies/" + str.Camel2Case(name) + ".go"
}
