package console

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
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
	m, err := supportconsole.NewMake(ctx, "job", ctx.Argument(0), support.Config.Paths.Job)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), m.GetSignature())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Job created successfully")

	if err = modify.GoFile(filepath.Join("app", "providers", "queue_service_provider.go")).
		Find(match.Imports()).Modify(modify.AddImport(m.GetPackageImportPath())).
		Find(match.Jobs()).Modify(modify.Register(fmt.Sprintf("&%s.%s{}", m.GetPackageName(), m.GetStructName()))).
		Apply(); err != nil {
		ctx.Warning(errors.QueueJobRegisterFailed.Args(err).Error())
		return nil
	}

	ctx.Success("Job registered successfully")

	return nil
}

func (r *JobMakeCommand) getStub() string {
	return JobStubs{}.Job()
}

// populateStub Populate the place-holders in the command stub.
func (r *JobMakeCommand) populateStub(stub string, packageName, structName, signature string) string {
	stub = strings.ReplaceAll(stub, "DummyJob", structName)
	stub = strings.ReplaceAll(stub, "DummySignature", str.Of(signature).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
