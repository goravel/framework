package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type AgentMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *AgentMakeCommand) Signature() string {
	return "make:agent"
}

// Description The console command description.
func (r *AgentMakeCommand) Description() string {
	return "Create a new agent"
}

// Extend The console command extend.
func (r *AgentMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the agent even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *AgentMakeCommand) Handle(ctx console.Context) error {
	make, err := supportconsole.NewMake(ctx, "agent", ctx.Argument(0), filepath.Join(support.Config.Paths.App, "agents"))
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(make.GetFilePath(), r.populateStub(r.getStub(), make.GetPackageName(), make.GetStructName())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Agent created successfully")

	return nil
}

func (r *AgentMakeCommand) getStub() string {
	return Stubs{}.Agent()
}

// populateStub Populate the place-holders in the command stub.
func (r *AgentMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyAgent", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
