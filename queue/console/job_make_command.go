package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type JobMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *JobMakeCommand) Signature() string {
	return "make:job"
}

// Description The console command description.
func (r *JobMakeCommand) Description() string {
	return "Create a new job class"
}

// Extend The console command extend.
func (r *JobMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the job even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *JobMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "job", ctx.Argument(0), filepath.Join("app", "jobs"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.Create(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Job created successfully")

	return nil
}

func (r *JobMakeCommand) getStub() string {
	return JobStubs{}.Job()
}

// populateStub Populate the place-holders in the command stub.
func (r *JobMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyJob", structName)
	stub = strings.ReplaceAll(stub, "DummyName", str.Of(structName).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
