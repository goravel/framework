package console

import (
	"github.com/goravel/framework/contracts/console/command"
)

type Command interface {
	// Signature set the unique signature for the command.
	Signature() string
	// Description the console command description.
	Description() string
	// Extend the console command extend.
	Extend() command.Extend
	// Handle execute the console command.
	Handle(ctx Context) error
}

//go:generate mockery --name=Context
type Context interface {
	// Argument get the value of a command argument.
	Argument(index int) string
	// Arguments get all the arguments passed to command.
	Arguments() []string
	// Option gets the value of a command option.
	Option(key string) string
	// OptionSlice looks up the value of a local StringSliceFlag, returns nil if not found
	OptionSlice(key string) []string
	// OptionBool looks up the value of a local BoolFlag, returns false if not found
	OptionBool(key string) bool
	// OptionFloat64 looks up the value of a local Float64Flag, returns zero if not found
	OptionFloat64(key string) float64
	// OptionFloat64Slice looks up the value of a local Float64SliceFlag, returns nil if not found
	OptionFloat64Slice(key string) []float64
	// OptionInt looks up the value of a local IntFlag, returns zero if not found
	OptionInt(key string) int
	// OptionIntSlice looks up the value of a local IntSliceFlag, returns nil if not found
	OptionIntSlice(key string) []int
	// OptionInt64 looks up the value of a local Int64Flag, returns zero if not found
	OptionInt64(key string) int64
	// OptionInt64Slice looks up the value of a local Int64SliceFlag, returns nil if not found
	OptionInt64Slice(key string) []int64
}
