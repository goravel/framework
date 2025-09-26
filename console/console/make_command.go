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
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type MakeCommand struct {
}

func NewMakeCommand() *MakeCommand {
	return &MakeCommand{}
}

// Signature The name and signature of the console command.
func (r *MakeCommand) Signature() string {
	return "make:command"
}

// Description The console command description.
func (r *MakeCommand) Description() string {
	return "Create a new Artisan command"
}

// Extend The console command extend.
func (r *MakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (r *MakeCommand) Handle(ctx console.Context) error {
	if err := r.initKernel(); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	return nil

	m, err := supportconsole.NewMake(ctx, "command", ctx.Argument(0), filepath.Join("app", "console", "commands"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), m.GetSignature())); err != nil {
		return err
	}

	ctx.Success("Console command created successfully")

	if err = modify.GoFile(filepath.Join("app", "console", "kernel.go")).
		Find(match.Imports()).Modify(modify.AddImport(m.GetPackageImportPath())).
		Find(match.Commands()).Modify(modify.Register(fmt.Sprintf("&%s.%s{}", m.GetPackageName(), m.GetStructName()))).
		Apply(); err != nil {
		ctx.Warning(errors.ConsoleCommandRegisterFailed.Args(err).Error())
		return nil
	}

	ctx.Success("Console command registered successfully")

	return nil
}

func (r *MakeCommand) getStub() string {
	return Stubs{}.Command()
}

func (r *MakeCommand) initKernel() error {
	kernelPath := filepath.Join("app", "console", "kernel.go")
	if file.Exists(kernelPath) {
		if !file.Contain(kernelPath, "func (kernel Kernel) Commands()") {
			if err := file.PutContent(kernelPath, Stubs{}.KernelCommands(), file.WithAppend()); err != nil {
				return err
			}

			if err := modify.GoFile(kernelPath).FindOrCreate(match.Imports(), modify.CreateImport).
				Modify(modify.AddImport("github.com/goravel/framework/contracts/console")).Apply(); err != nil {
				return err
			}
		}
	} else {
		if err := file.PutContent(kernelPath, Stubs{}.Kernel()); err != nil {
			return err
		}
	}

	return nil
}

// populateStub Populate the place-holders in the command stub.
func (r *MakeCommand) populateStub(stub string, packageName, structName, signature string) string {
	stub = strings.ReplaceAll(stub, "DummyCommand", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)
	stub = strings.ReplaceAll(stub, "DummySignature", str.Of(signature).Kebab().Prepend("app:").String())

	return stub
}
