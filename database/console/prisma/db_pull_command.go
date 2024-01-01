package prisma

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type DBPullCommand struct{}

func NewDBPullCommand() *DBPullCommand {
	return &DBPullCommand{}
}

// Signature The name and signature of the console command.
func (receiver *DBPullCommand) Signature() string {
	return "prisma:db:pull"
}

// Description The console command description.
func (receiver *DBPullCommand) Description() string {
	return "Pull the state from the database to the Prisma schema using introspection"
}

// Extend The console command extend.
func (receiver *DBPullCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command
func (r *DBPullCommand) Handle(ctx console.Context) error {
	args := ctx.Argument(0)
	cliCmd := append([]string{"db", "pull"}, strings.Split(args, " ")...)
	return cli.Run(cliCmd, true)
}
