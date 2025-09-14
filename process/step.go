package process

import contractsprocess "github.com/goravel/framework/contracts/process"

var _ contractsprocess.Step = (*Step)(nil)

type Step struct {
	key string
	name string
	args []string
}

func NewStep(key, name string, args []string) *Step {
	return &Step{
		key: key,
		name: name,
		args: args,
	}
}

func (r *Step) As(key string) contractsprocess.Step {
	r.key = key
	return r
}

