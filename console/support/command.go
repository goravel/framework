package support

import "github.com/urfave/cli/v2"

type Command interface {
	//Signature The name and signature of the console command.
	Signature() string
	//Description The console command description.
	Description() string
	//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
	Flags() []cli.Flag
	//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
	Subcommands() []*cli.Command
	//Handle Execute the console command.
	Handle(c *cli.Context) error
}
