package prisma

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type InitCommand struct {
}

func NewInitCommand() *InitCommand {
	return &InitCommand{}
}

// Signature The name and signature of the console command.
func (receiver *InitCommand) Signature() string {
	return "prisma:init"
}

// Description The console command description.
func (receiver *InitCommand) Description() string {
	return "Set up a new Prisma project"
}

// Extend The console command extend.
func (receiver *InitCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command.
func (receiver *InitCommand) Handle(ctx console.Context) error {
	cliCmds := []string{
		"init", "--generator-provider", ".",
		"--datasource-provider", "sqlite",
		"--url", "file:dev.db",
	}
	return cli.Run(cliCmds, true)
}
