package prisma

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type MigrateDevCommand struct{}

func NewMigrateDevCommand() *MigrateDevCommand {
	return &MigrateDevCommand{}
}

// Signature The name and signature of the console command.
func (receiver *MigrateDevCommand) Signature() string {
	return "prisma:migrate:dev"
}

// Description The console command description.
func (receiver *MigrateDevCommand) Description() string {
	return "🏋️  Create a migration from changes in Prisma schema, apply it to the database, trigger generators (e.g. Prisma Client)"
}

// Extend The console command extend.
func (receiver *MigrateDevCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command.
func (receiver *MigrateDevCommand) Handle(ctx console.Context) error {
	args := strings.Split(ctx.Argument(0), " ")
	cliCmds := []string{
		"migrate", "dev",
		"--name", "init",
	}
	cliCmds = append(cliCmds, args...)
	return cli.Run(cliCmds, true)
}
