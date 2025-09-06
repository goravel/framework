package process

import (
	"context"
	"io"
	"time"
)

type OnPipeOutputFunc func(key string, typ OutputType, line []byte)

type Pipeline interface {
	DisableBuffering() Pipeline
	Env(vars map[string]string) Pipeline
	Input(in io.Reader) Pipeline
	Path(path string) Pipeline
	Quietly() Pipeline
	OnOutput(handler OnPipeOutputFunc) Pipeline
	Run(func(builder Pipe)) (Result, error)
	Start(func(builder Pipe)) (RunningPipe, error)
	Timeout(timeout time.Duration) Pipeline
	WithContext(ctx context.Context) Pipeline
}

type Pipe interface {
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
