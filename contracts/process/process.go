package process

import (
	"context"
	"io"
	"os"
	"time"
)

type Command interface {
	Path(dir string) Command
	Env(env map[string]string) Command
	Input(reader io.Reader) Command
	Timeout(duration time.Duration) Command
	IdleTimeout(duration time.Duration) Command
	Quietly() Command
	Tty() Command
	OnOutput(handler func(typ, line string)) Command
	Run(ctx context.Context) (Result, error)
	Start(ctx context.Context) (Running, error)
}

type Result interface {
	Successful() bool
	Failed() bool
	ExitCode() int
	Output() string
	ErrorOutput() string
	Command() string
	Duration() time.Duration
	ProcessState() *os.ProcessState
	SeeInOutput(needle string) bool
}

type Running interface {
	PID() int
	Command() string
	Running() bool
	Output() string
	ErrorOutput() string
	Wait() Result
	Kill() error
	Signal(sig os.Signal) error
	Process() *os.Process
}
