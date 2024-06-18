package console

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
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
	name, err := supportconsole.GetName(ctx, "mail", ctx.Argument(0), receiver.getPath)
	if err != nil {
		color.Red().Println(err)
		return nil
	}

	if err := file.Create(receiver.getPath(name), receiver.populateStub(receiver.getStub(), name)); err != nil {
		return err
	}

	color.Green().Println("Mail created successfully")

	return nil
}

func (receiver *MailMakeCommand) getStub() string {
	return Stubs{}.Mail()
}

// populateStub Populate the place-holders in the command stub.
func (receiver *MailMakeCommand) populateStub(stub string, name string) string {
	modelName, packageName, _ := receiver.parseName(name)

	stub = strings.ReplaceAll(stub, "DummyMail", str.Case2Camel(modelName))
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}

// getPath Get the full path to the command.
func (receiver *MailMakeCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	modelName, _, folderPath := receiver.parseName(name)

	return filepath.Join(pwd, "app", "mails", folderPath, str.Camel2Case(modelName)+".go")
}

// parseName Parse the name to get the model name, package name and folder path.
func (receiver *MailMakeCommand) parseName(name string) (string, string, string) {
	name = strings.TrimSuffix(name, ".go")

	segments := strings.Split(name, "/")

	modelName := segments[len(segments)-1]

	packageName := "mails"
	folderPath := ""

	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
		packageName = segments[len(segments)-2]
	}

	return modelName, packageName, folderPath
}
