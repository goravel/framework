package console

import (
	"os"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type SeederMakeCommand struct {
}

func NewSeederMakeCommand() *SeederMakeCommand {
	return &SeederMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *SeederMakeCommand) Signature() string {
	return "make:seeder"
}

// Description The console command description.
func (receiver *SeederMakeCommand) Description() string {
	return "Create a new seeder class"
}

// Extend The console command extend.
func (receiver *SeederMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *SeederMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.getStub(name)); err != nil {
		return err
	}

	color.Greenln("Seeder created successfully")

	return nil
}

func (receiver *SeederMakeCommand) getStub(name string) string {
	return Stubs{}.Seeder(name)
}

// getPath Get the full path to the command.
func (receiver *SeederMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/database/seeders/" + str.Camel2Case(name) + ".go"
}
