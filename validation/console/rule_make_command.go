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

type RuleMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *RuleMakeCommand) Signature() string {
	return "make:rule"
}

//Description The console command description.
func (receiver *RuleMakeCommand) Description() string {
	return "Create a new rule class"
}

//Extend The console command extend.
func (receiver *RuleMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *RuleMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Rule created successfully")

	return nil
}

func (receiver *RuleMakeCommand) getStub() string {
	return Stubs{}.Request()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *RuleMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyRule", str.Case2Camel(name))
	stub = strings.ReplaceAll(stub, "DummyName", str.Camel2Case(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *RuleMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/rules/" + str.Camel2Case(name) + ".go"
}
