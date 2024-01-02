package prisma

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type VersionCommand struct{}

func NewVersionCommand() *VersionCommand {
	return &VersionCommand{}
}

// Signature The name and signature of the console command.
func (receiver *VersionCommand) Signature() string {
	return "prisma:version"
}

// Description The console command description.
func (receiver *VersionCommand) Description() string {
	return "Print current version of Prisma components"
}

// Extend The console command extend.
func (receiver *VersionCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command.
func (receiver *VersionCommand) Handle(ctx console.Context) error {
	args := strings.Split(ctx.Argument(0), " ")
	cliCmds := []string{
		"version",
	}
	cliCmds = append(cliCmds, args...)
	return cli.Run(cliCmds, true)
}
