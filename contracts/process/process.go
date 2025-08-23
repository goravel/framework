package process

import (
	"io"
	"os"
	"time"
)

type SingleBuilder interface {
	Path(dir string) SingleBuilder
	Env(env map[string]string) SingleBuilder
	Input(reader io.Reader) SingleBuilder
	Timeout(duration time.Duration) SingleBuilder
	IdleTimeout(duration time.Duration) SingleBuilder
	Quietly() SingleBuilder
	Tty() SingleBuilder
	OnOutput(handler func(typ, line string)) SingleBuilder
	Run() (Result, error)
	Start() (Running, error)
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
