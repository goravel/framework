package process

import (
	"os"
	"time"
)

type RunningPool interface {
	PIDs() map[string]int
	Running() bool
	Done() <-chan struct{}
	Wait() map[string]Result
	Stop(timeout time.Duration, sig ...os.Signal) error
	Signal(sig os.Signal) error
}
