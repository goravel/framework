package console

import (
	"github.com/urfave/cli/v3"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type HelpCommand struct {
}

func NewHelpCommand() *HelpCommand {
	return &HelpCommand{}
}

// Signature The name and signature of the console command.
func (r *HelpCommand) Signature() string {
	return "help"
}

// Description The console command description.
func (r *HelpCommand) Description() string {
	return "Shows a list of commands"
}

// Extend The console command extend.
func (r *HelpCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (r *HelpCommand) Handle(ctx console.Context) error {
	return cli.ShowAppHelp(ctx.Instance())
}
