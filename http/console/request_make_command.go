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

type RequestMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *RequestMakeCommand) Signature() string {
	return "make:request"
}

// Description The console command description.
func (receiver *RequestMakeCommand) Description() string {
	return "Create a new request class"
}

// Extend The console command extend.
func (receiver *RequestMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the request even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *RequestMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "request", ctx.Argument(0), filepath.Join("app", "http", "requests"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Request created successfully")

	return nil
}

func (receiver *RequestMakeCommand) getStub() string {
	return Stubs{}.Request()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *RequestMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyRequest", structName)
	stub = strings.ReplaceAll(stub, "DummyField", "Name string `form:\"name\" json:\"name\"`")
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
