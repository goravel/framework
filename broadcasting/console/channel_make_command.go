package console

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
)

type ChannelMakeCommand struct{}

func NewChannelMakeCommand() *ChannelMakeCommand {
	return &ChannelMakeCommand{}
}

func (r *ChannelMakeCommand) Signature() string {
	return "make:channel"
}

func (r *ChannelMakeCommand) Description() string {
	return "Create a new channel class for broadcasting authorization"
}

func (r *ChannelMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the channel even if it already exists",
			},
		},
	}
}

func (r *ChannelMakeCommand) Handle(ctx console.Context) error {
	for _, name := range supportconsole.MakeNames(ctx) {
		if err := r.makeOne(ctx, name); err != nil {
			ctx.Error(err.Error())
		}
	}

	return nil
}

func (r *ChannelMakeCommand) makeOne(ctx console.Context, name string) error {
	m, err := supportconsole.NewMake(ctx, "channel", name, support.Config.Paths.Broadcasting)
	if err != nil {
		return err
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Channel created successfully")

	return nil
}

func (r *ChannelMakeCommand) getStub() string {
	return Stubs{}.Channel()
}

func (r *ChannelMakeCommand) populateStub(stub string, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyChannel", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
