package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
)

type ListCommand struct {
	App foundation.Application
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
func (receiver *ListCommand) Extend() command.Extend {
	return command.Extend{}
}

//Handle Execute the console command.
func (receiver *ListCommand) Handle(ctx console.Context) error {
	receiver.App.MakeArtisan().Call("--help")

	return nil
}
