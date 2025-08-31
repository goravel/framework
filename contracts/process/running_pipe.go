package process

import (
	"os"
	"time"
)

type RunningPipe interface {
	PIDs() map[string]int
	Running() bool
	Done() <-chan struct{}
	Wait() Result
	Stop(timeout time.Duration, sig ...os.Signal) error
	Signal(sig os.Signal) error
}
