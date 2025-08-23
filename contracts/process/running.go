package process

import (
	"os"
	"time"
)

type Running interface {
	PID() int
	Running() bool
	Output() string
	ErrorOutput() string
	LatestOutput() string
	LatestErrorOutput() string
	Wait() Result
	Stop(timeout time.Duration, sig ...os.Signal) error
	Signal(sig os.Signal) error
}
