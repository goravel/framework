package console

import (
	"github.com/goravel/framework/contracts/console/command"
)

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

//go:generate mockery --name=Context
type Context interface {
	Argument(index int) string
	Arguments() []string
	Option(key string) string
}
