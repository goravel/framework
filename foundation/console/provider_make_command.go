package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type ProviderMakeCommand struct {
}

func NewProviderMakeCommand() *ProviderMakeCommand {
	return &ProviderMakeCommand{}
}

// Signature The name and signature of the console command.
func (r *ProviderMakeCommand) Signature() string {
	return "make:provider"
}

// Description The console command description.
func (r *ProviderMakeCommand) Description() string {
	return "Create a new service provider class"
}

// Extend The console command extend.
func (r *ProviderMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the provider even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *ProviderMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "provider", ctx.Argument(0), filepath.Join("app", "providers"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	stub := r.getStub()

	if err := file.PutContent(m.GetFilePath(), r.populateStub(stub, m.GetPackageName(), m.GetStructName())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Provider created successfully")

	return nil
}

func (r *ProviderMakeCommand) getStub() string {
	return Stubs{}.ServiceProvider()
}

// populateStub Populate the place-holders in the command stub.
func (r *ProviderMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyServiceProvider", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
