package prisma

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type MigrateDeployCommand struct {
}

func NewMigrateDeployCommand() *MigrateDeployCommand {
	return &MigrateDeployCommand{}
}

// Signature The name and signature of the console command.
func (receiver *MigrateDeployCommand) Signature() string {
	return "prisma:migrate:deploy"
}

// Description The console command description.
func (receiver *MigrateDeployCommand) Description() string {
	return "Apply pending migrations to update the database schema in production/staging"
}

// Extend The console command extend.
func (receiver *MigrateDeployCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command.
func (receiver *MigrateDeployCommand) Handle(ctx console.Context) error {
	args := strings.Split(ctx.Argument(0), " ")
	cliCmds := []string{
		"migrate", "deploy",
	}
	cliCmds = append(cliCmds, args...)
	return cli.Run(cliCmds, true)
}
