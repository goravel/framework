package console

import (
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type ClearCommand struct {
	cache cache.Cache
}

func NewClearCommand(cache cache.Cache) *ClearCommand {
	return &ClearCommand{cache: cache}
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
	if receiver.cache.Flush() {
		color.Greenln("Application cache cleared")
	} else {
		color.Redln("Clear Application cache Failed")
	}

	return nil
}
