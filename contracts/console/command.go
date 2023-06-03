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
	OptionSlice(key string) []string
	OptionBool(key string) bool
	OptionFloat64(key string) float64
	OptionFloat64Slice(key string) []float64
	OptionInt(key string) int
	OptionIntSlice(key string) []int
	OptionInt64(key string) int64
	OptionInt64Slice(key string) []int64
}
