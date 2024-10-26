package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
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
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the middleware even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *MiddlewareMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "middleware", ctx.Argument(0), filepath.Join("app", "http", "middleware"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Middleware created successfully")

	return nil
}

func (receiver *MiddlewareMakeCommand) getStub() string {
	return Stubs{}.Middleware()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *MiddlewareMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyMiddleware", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
