package console

import (
	"errors"
	"os"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type MiddlewareMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *MiddlewareMakeCommand) Signature() string {
	return "make:middleware"
}

// Description The console command description.
func (receiver *MiddlewareMakeCommand) Description() string {
	return "Create a new middleware class"
}

// Extend The console command extend.
func (receiver *MiddlewareMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
	}
}

// Handle Execute the console command.
func (receiver *MiddlewareMakeCommand) Handle(ctx console.Context) error {
	name := ctx.Argument(0)
	if name == "" {
		return errors.New("Not enough arguments (missing: name) ")
	}

	file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name))
	color.Greenln("Middleware created successfully")

	return nil
}

func (receiver *MiddlewareMakeCommand) getStub() string {
	return Stubs{}.Middleware()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *MiddlewareMakeCommand) populateStub(stub string, name string) string {
	stub = strings.ReplaceAll(stub, "DummyMiddleware", str.Case2Camel(name))

	return stub
}

// getPath Get the full path to the command.
func (receiver *MiddlewareMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return pwd + "/app/http/middleware/" + str.Camel2Case(name) + ".go"
}
