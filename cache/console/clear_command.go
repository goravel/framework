package console

import (
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
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
func (receiver *ClearCommand) Extend() command.Extend {
	return command.Extend{
		Category: "cache",
	}
}

//Handle Execute the console command.
func (receiver *ClearCommand) Handle(ctx console.Context) error {
	res := facades.Cache.Flush()

	if res {
		color.Greenln("Application cache cleared")
	} else {
		color.Redln("Clear Application cache Failed")
	}

	return nil
}
