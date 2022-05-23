package stubs

type ConsoleStubs struct {
}

//Command Create a command.
func (receiver ConsoleStubs) Command() string {
	return `package commands

import (
	"github.com/goravel/framework/contracts/console"
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

//Extend The console command extend.
func (receiver *DummyCommand) Extend() console.CommandExtend {
	return console.CommandExtend{}
}

//Handle Execute the console command.
func (receiver *DummyCommand) Handle(c *cli.Context) error {
	
	return nil
}
`
}
