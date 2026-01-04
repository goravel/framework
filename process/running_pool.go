package process

import (
	"context"
	"os"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.RunningPool = (*RunningPool)(nil)

type RunningPool struct {
	running []contractsprocess.Running
	keys    []string
	cancel  context.CancelFunc
	results map[string]contractsprocess.Result
	done    chan struct{}
}

func NewRunningPool(
	running []contractsprocess.Running,
	keys []string,
	cancel context.CancelFunc,
	results map[string]contractsprocess.Result,
	done chan struct{},
) *RunningPool {
	return &RunningPool{
		running: running,
		keys:    keys,
		cancel:  cancel,
		results: results,
		done:    done,
	}
}

func (r *RunningPool) PIDs() map[string]int {
	m := make(map[string]int, len(r.running))
	for i, proc := range r.running {
		pid := 0
		if proc != nil {
			pid = proc.PID()
		}
		m[r.keys[i]] = pid
	}
	return m
}

func (r *RunningPool) Running() bool {
	select {
	case <-r.done:
		return false
	default:
		return true
	}
}

func (r *RunningPool) Done() <-chan struct{} {
	return r.done
}

func (r *RunningPool) Wait() map[string]contractsprocess.Result {
	<-r.Done()
	return r.results
}

func (r *RunningPool) Stop(timeout time.Duration, sig ...os.Signal) error {
	var firstErr error
	for _, proc := range r.running {
		if proc == nil {
			continue
		}
		if err := proc.Stop(timeout, sig...); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *RunningPool) Signal(sig os.Signal) error {
	var firstErr error
	for _, proc := range r.running {
		if proc == nil {
			continue
		}
		if err := proc.Signal(sig); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
