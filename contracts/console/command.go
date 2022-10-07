package console

import (
	"github.com/goravel/framework/contracts/console/command"
)

//go:generate mockery --name=Command
type Command interface {
	//Signature The name and signature of the console command.
	Signature() string
	//Description The console command description.
	Description() string
	//Extend The console command extend.
	Extend() command.Extend
	//Handle Execute the console command.
	Handle(ctx Context) error
}

type Context interface {
	Argument(index int) string
	Arguments() []string
	Option(key string) string
}
