package console

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type TestMakeCommand struct {
}

func NewTestMakeCommand() *TestMakeCommand {
	return &TestMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *TestMakeCommand) Signature() string {
	return "make:test"
}

// Description The console command description.
func (receiver *TestMakeCommand) Description() string {
	return "Create a new test class"
}

// Extend The console command extend.
func (receiver *TestMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the test even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *TestMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "test", ctx.Argument(0), "tests")
	if err != nil {
		color.Errorln(err)
		return nil
	}

	stub := receiver.getStub()

	if err := file.Create(m.GetFilePath(), receiver.populateStub(stub, m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Test created successfully")

	return nil
}

func (receiver *TestMakeCommand) getStub() string {
	return Stubs{}.Test()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *TestMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyTest", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
