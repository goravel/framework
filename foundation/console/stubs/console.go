package stubs

type ConsoleStubs struct {
}

//Command Create a command.
func (receiver ConsoleStubs) Command() string {
	return `package commands

import (
	"github.com/urfave/cli/v2"
)

type DummyCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *DummyCommand) Signature() string {
	return "command:name"
}

//Description The console command description.
func (receiver *DummyCommand) Description() string {
	return "Command description"
}

//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
func (receiver *DummyCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
func (receiver *DummyCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

//Handle Execute the console command.
func (receiver *DummyCommand) Handle(c *cli.Context) error {
	
	return nil
}
`
}
