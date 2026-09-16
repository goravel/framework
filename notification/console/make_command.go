package console

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type NotificationMakeCommand struct{}

func NewNotificationMakeCommand() *NotificationMakeCommand {
	return &NotificationMakeCommand{}
}

func (r *NotificationMakeCommand) Signature() string {
	return "make:notification"
}

func (r *NotificationMakeCommand) Description() string {
	return "Create a new notification class"
}

func (r *NotificationMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the notification even if it already exists",
			},
			&command.BoolFlag{
				Name:  "database",
				Usage: "Generate a notification template for the database channel instead of mail",
			},
		},
	}
}

func (r *NotificationMakeCommand) Handle(ctx console.Context) error {
	for _, name := range supportconsole.MakeNames(ctx) {
		if err := r.makeOne(ctx, name); err != nil {
			ctx.Error(err.Error())
		}
	}

	return nil
}

func (r *NotificationMakeCommand) makeOne(ctx console.Context, name string) error {
	m, err := supportconsole.NewMake(ctx, "notification", name, support.Config.Paths.Notifications)
	if err != nil {
		return err
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(ctx), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Notification created successfully")

	return nil
}

func (r *NotificationMakeCommand) getStub(ctx console.Context) string {
	if ctx.OptionBool("database") {
		return Stubs{}.DatabaseNotification()
	}
	return Stubs{}.MailNotification()
}

func (r *NotificationMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyNotification", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
