package console

import (
	"errors"
	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/queue/console/stubs"
	"github.com/goravel/framework/support"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

type JobMakeCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *JobMakeCommand) Signature() string {
	return "make:job"
}

//Description The console command description.
func (receiver *JobMakeCommand) Description() string {
	return "Create a new job class"
}

//Extend The console command extend.
func (receiver *JobMakeCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "make",
	}
}

//Handle Execute the console command.
func (receiver *JobMakeCommand) Handle(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	support.Helpers{}.CreateFile(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Job created successfully")

	return nil
}

func (receiver *JobMakeCommand) getStub() string {
	return stubs.JobStubs{}.Job()
}

//populateStub Populate the place-holders in the command stub.
func (receiver *JobMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyJob", support.Helpers{}.Case2Camel(name))
	stub = strings.ReplaceAll(stub, "DummyName", support.Helpers{}.Camel2Case(name))

	return stub
}

//getPath Get the full path to the command.
func (receiver *JobMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/jobs/" + support.Helpers{}.Camel2Case(name) + ".go"
}
