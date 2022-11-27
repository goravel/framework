package console

import "github.com/urfave/cli/v2"

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
