package process

import (
	"context"
	"io"
	"time"
)

type OnPipeOutputFunc func(key string, typ OutputType, line []byte)

type Pipe interface {
	DisableBuffering() Pipe
	Env(vars map[string]string) Pipe
	Input(in io.Reader) Pipe
	Path(path string) Pipe
	Quietly() Pipe
	OnOutput(handler OnPipeOutputFunc) Pipe
	Run(func(builder PipeBuilder)) (Result, error)
	Start(func(builder PipeBuilder)) (RunningPipe, error)
	Timeout(timeout time.Duration) Pipe
	WithContext(ctx context.Context) Pipe
}

type PipeBuilder interface {
	Command(name string, arg ...string) *Step
}

type Step struct {
	Key  string
	Name string
	Args []string
}

func (s *Step) As(key string) *Step {
	s.Key = key
	return s
}
