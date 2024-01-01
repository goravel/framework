package prisma

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type DBSeedCommand struct{}

func NewDBSeedCommand() *DBSeedCommand {
	return &DBSeedCommand{}
}

// Signature The name and signature of the console command.
func (receiver *DBSeedCommand) Signature() string {
	return "prisma:db:seed"
}

// Description The console command description.
func (receiver *DBSeedCommand) Description() string {
	return "🙌  Seed your database"
}

// Extend The console command extend.
func (receiver *DBSeedCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command
func (r *DBSeedCommand) Handle(ctx console.Context) error {
	args := ctx.Argument(0)
	cliCmd := append([]string{"db", "push"}, strings.Split(args, " ")...)
	return cli.Run(cliCmd, true)

}
