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
	Command(name string, arg ...string) Process
	Env(vars map[string]string) Process
	Forever() Process
	IdleTimeout(timeout time.Duration) Process
	Input(in io.Reader) Process
	Path(path string) Process
	Quietly() Process
	OnOutput(handler OnOutputFunc) Process
	Run(ctx context.Context) (Result, error)
	Start(ctx context.Context) (Running, error)
	Timeout(timeout time.Duration) Process
	TTY() Process
}
