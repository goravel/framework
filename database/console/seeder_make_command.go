package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type SeederMakeCommand struct {
}

func NewSeederMakeCommand() *SeederMakeCommand {
	return &SeederMakeCommand{}
}

// Signature The name and signature of the console command.
func (r *SeederMakeCommand) Signature() string {
	return "make:seeder"
}

// Description The console command description.
func (r *SeederMakeCommand) Description() string {
	return "Create a new seeder class"
}

// Extend The console command extend.
func (r *SeederMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the seeder even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *SeederMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "seeder", ctx.Argument(0), filepath.Join("database", "seeders"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.Create(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Seeder created successfully")

	return nil
}

func (r *SeederMakeCommand) getStub() string {
	return Stubs{}.Seeder()
}

// populateStub Populate the place-holders in the command stub.
func (r *SeederMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummySeeder", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
