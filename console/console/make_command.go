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
	m, err := supportconsole.NewMake(ctx, "command", ctx.Argument(0), filepath.Join("app", "console", "commands"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.Create(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Console command created successfully")

	return nil
}

func (r *MakeCommand) getStub() string {
	return Stubs{}.Command()
}

// populateStub Populate the place-holders in the command stub.
func (r *MakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyCommand", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)
	stub = strings.ReplaceAll(stub, "DummySignature", str.Of(structName).Kebab().Prepend("app:").String())

	return stub
}
