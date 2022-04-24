package console

import (
	"errors"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation/console/stubs"
	"github.com/goravel/framework/support"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

type ConsoleMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *ConsoleMakeCommand) Signature() string {
	return "make:command"
}

//Description The console command description.
func (receiver *ConsoleMakeCommand) Description() string {
	return "Create a new Artisan command"
}

//Extend The console command extend.
func (receiver *ConsoleMakeCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *ConsoleMakeCommand) Handle(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	support.Helpers{}.CreateFile(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))

	return nil
}

func (receiver *ConsoleMakeCommand) getStub() string {
	return stubs.ConsoleStubs{}.Command()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *ConsoleMakeCommand) populateStub(stub string, name string) string {
	return strings.ReplaceAll(stub, "DummyCommand", support.Helpers{}.Case2Camel(name))
}

//getPath Get the full path to the command.
func (receiver *ConsoleMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/console/commands/" + support.Helpers{}.Camel2Case(name) + ".go"
}
