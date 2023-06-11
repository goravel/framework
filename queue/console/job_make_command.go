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

type JobMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *JobMakeCommand) Signature() string {
	return "make:job"
}

// Description The console command description.
func (receiver *JobMakeCommand) Description() string {
	return "Create a new job class"
}

// Extend The console command extend.
func (receiver *JobMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *JobMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Greenln("Job created successfully")

	return nil
}

func (receiver *JobMakeCommand) getStub() string {
	return JobStubs{}.Job()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *JobMakeCommand) populateStub(stub string, name string) string {
	jobName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyJob", str.Case2Camel(jobName))
	stub = strings.ReplaceAll(stub, "DummyName", str.Camel2Case(jobName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *JobMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	jobName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "jobs", folderPath, str.Camel2Case(jobName)+".go")
}

// parseName Parse the name to get the job name, package name and folder path.
func (receiver *JobMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	jobName := segments[len(segments)-1]

	packageName := "jobs"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return jobName, packageName, folderPath
}
