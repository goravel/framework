package console

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
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

//Extend The console command extend.
func (receiver *ClearCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "cache",
	}
}

//Handle Execute the console command.
func (receiver *ClearCommand) Handle(c *cli.Context) error {
	res := facades.Cache.Flush()

	if res {
		color.Greenln("Application cache cleared")
	} else {
		color.Redln("Clear Application cache Failed")
	}

	return nil
}
