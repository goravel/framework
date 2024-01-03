package prisma

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type FormatCommand struct{}

func NewFormatCommand() *FormatCommand {
	return &FormatCommand{}
}

// Signature The name and signature of the console command.
func (receiver *FormatCommand) Signature() string {
	return "prisma:format"
}

// Description The console command description.
func (receiver *FormatCommand) Description() string {
	return "Format a Prisma schema"
}

// Extend The console command extend.
func (receiver *FormatCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command
func (r *FormatCommand) Handle(ctx console.Context) error {
	args := ctx.Argument(0)
	cliCmds := []string{"debug"}
	cliCmds = append(cliCmds, strings.Split(args, " ")...)
	return cli.Run(cliCmds, true)
}
