package console

import (
	"fmt"
	"github.com/goravel/framework/support/facades"
	"github.com/urfave/cli/v2"
)

type ClearCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *ClearCommand) Signature() string {
	return "cache:clear"
}

//Description The console command description.
func (receiver *ClearCommand) Description() string {
	return "Flush the application cache"
}

//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
func (receiver *ClearCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
func (receiver *ClearCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

//Handle Execute the console command.
func (receiver *ClearCommand) Handle(c *cli.Context) error {
	res := facades.Cache.Flush()

	if res {
		fmt.Println("Application cache cleared")
	} else {
		fmt.Println("Clear Application cache Failed")
	}

	return nil
}
