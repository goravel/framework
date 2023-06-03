package console

import (
	"github.com/urfave/cli/v2"
)

type CliContext struct {
	instance *cli.Context
}

func (r *CliContext) Argument(index int) string {
	return r.instance.Args().Get(index)
}

func (r *CliContext) Arguments() []string {
	return r.instance.Args().Slice()
}

func (r *CliContext) Option(key string) string {
	return r.instance.String(key)
}

func (r *CliContext) OptionSlice(key string) []string {
	return r.instance.StringSlice(key)
}

func (r *CliContext) OptionBool(key string) bool {
	return r.instance.Bool(key)
}

func (r *CliContext) OptionFloat64(key string) float64 {
	return r.instance.Float64(key)
}

func (r *CliContext) OptionFloat64Slice(key string) []float64 {
	return r.instance.Float64Slice(key)
}

func (r *CliContext) OptionInt(key string) int {
	return r.instance.Int(key)
}

func (r *CliContext) OptionIntSlice(key string) []int {
	return r.instance.IntSlice(key)
}

func (r *CliContext) OptionInt64(key string) int64 {
	return r.instance.Int64(key)
}

func (r *CliContext) OptionInt64Slice(key string) []int64 {
	return r.instance.Int64Slice(key)
}
