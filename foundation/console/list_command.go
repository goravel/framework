package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/facades"
	"github.com/urfave/cli/v2"
)

type ListCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *ListCommand) Signature() string {
	return "list"
}

//Description The console command description.
func (receiver *ListCommand) Description() string {
	return "List commands"
}

//Extend The console command extend.
func (receiver *ListCommand) Extend() console.CommandExtend {
	return console.CommandExtend{}
}

//Handle Execute the console command.
func (receiver *ListCommand) Handle(c *cli.Context) error {
	facades.Artisan.Call("--help")

	return nil
}
