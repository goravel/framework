package process

import (
	"os"
	"time"
)

type RunningPool interface {
	Done() <-chan struct{}
	PIDs() map[string]int
	Running() bool
	Signal(sig os.Signal) error
	Stop(timeout time.Duration, sig ...os.Signal) error
	Wait() map[string]Result
}
