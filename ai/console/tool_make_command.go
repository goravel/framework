package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type ToolMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *ToolMakeCommand) Signature() string {
	return "make:tool"
}

// Description The console command description.
func (r *ToolMakeCommand) Description() string {
	return "Create a new agent tool"
}

// Extend The console command extend.
func (r *ToolMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the tool even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *ToolMakeCommand) Handle(ctx console.Context) error {
	make, err := supportconsole.NewMake(ctx, "tool", ctx.Argument(0), filepath.Join(support.Config.Paths.App, "tools"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(make.GetFilePath(), r.populateStub(r.getStub(), make.GetPackageName(), make.GetStructName())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Tool created successfully")

	return nil
}

func (r *ToolMakeCommand) getStub() string {
	return Stubs{}.Tool()
}

// populateStub Populate the place-holders in the command stub.
func (r *ToolMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyTool", structName)
	stub = strings.ReplaceAll(stub, "DummyName", str.Of(structName).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
