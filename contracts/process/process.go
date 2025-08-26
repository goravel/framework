package process

import (
	"context"
	"io"
	"time"
)

type OutputType int

const (
	OutputTypeStdout OutputType = iota
	OutputTypeStderr
)

type OnOutputFunc func(typ OutputType, line []byte)

type Process interface {
	Env(vars map[string]string) Process
	Input(in io.Reader) Process
	Path(path string) Process
	Quietly() Process
	OnOutput(handler OnOutputFunc) Process
	Run(name string, arg ...string) (Result, error)
	Start(name string, arg ...string) (Running, error)
	Timeout(timeout time.Duration) Process
	TTY() Process
	WithContext(ctx context.Context) Process
}
