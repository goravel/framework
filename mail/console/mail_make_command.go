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

type MailMakeCommand struct {
}

func NewMailMakeCommand() *MailMakeCommand {
	return &MailMakeCommand{}
}

// Signature The name and signature of the console command.
func (receiver *MailMakeCommand) Signature() string {
	return "make:mail"
}

// Description The console command description.
func (receiver *MailMakeCommand) Description() string {
	return "Create a new mail class"
}

// Extend The console command extend.
func (receiver *MailMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the mail even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *MailMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "mail", ctx.Argument(0), filepath.Join("app", "mails"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if err := file.Create(m.GetFilePath(), receiver.populateStub(receiver.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	color.Successln("Mail created successfully")

	return nil
}

func (receiver *MailMakeCommand) getStub() string {
	return Stubs{}.Mail()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *MailMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyMail", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
