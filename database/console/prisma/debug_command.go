package prisma

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/steebchen/prisma-client-go/cli"
)

type DebugCommand struct{}

func NewDebugCommand() *DebugCommand {
	return &DebugCommand{}
}

// Signature The name and signature of the console command.
func (receiver *DebugCommand) Signature() string {
	return "prisma:debug"
}

// Description The console command description.
func (receiver *DebugCommand) Description() string {
	return "Print information helpful for debugging and bug reports"
}

// Extend The console command extend.
func (receiver *DebugCommand) Extend() command.Extend {
	return command.Extend{
		Category: "prisma",
	}
}

// Handle Execute the console command
func (r *DebugCommand) Handle(ctx console.Context) error {
	return cli.Run([]string{"debug"}, true)
}
