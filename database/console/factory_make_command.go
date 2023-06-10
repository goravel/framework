package console

import (
	"log"
	"os"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type FactoryMakeCommand struct {
}

func NewFactoryMakeCommand() *FactoryMakeCommand {
	return &FactoryMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *FactoryMakeCommand) Signature() string {
	return "make:factory"
}

// Description The console command description.
func (receiver *FactoryMakeCommand) Description() string {
	return "Create a new model factory"
}

// Extend The console command extend.
func (receiver *FactoryMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *FactoryMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		color.Redln("Not enough arguments (missing: name)")

		return nil
	}
	log.Printf("name: %s", name)
	log.Println("receiver.getPath(name): ", receiver.getPath(name))
	log.Println("receiver.getStub(name): ", receiver.getStub(name))
	if err := file.Create(receiver.getPath(name), receiver.getStub(name)); err != nil {
		return err
	}

	color.Greenln("Factory created successfully")

	return nil
}

func (receiver *FactoryMakeCommand) getStub(name string) string {
	return Stubs{}.Factory(name)
}

// getPath Get the full path to the command.
func (receiver *FactoryMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/database/factories/" + str.Camel2Case(name) + ".go"
}
