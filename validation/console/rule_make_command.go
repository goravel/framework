package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type RuleMakeCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *RuleMakeCommand) Signature() string {
	return "make:rule"
}

// Description The console command description.
func (receiver *RuleMakeCommand) Description() string {
	return "Create a new rule class"
}

// Extend The console command extend.
func (receiver *RuleMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the rule even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *RuleMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "rule", ctx.Argument(0), filepath.Join("app", "rules"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Rule created successfully")

	return nil
}

func (receiver *RuleMakeCommand) getStub() string {
	return Stubs{}.Rule()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *RuleMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyRule", structName)
	stub = strings.ReplaceAll(stub, "DummyName", str.Of(structName).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
