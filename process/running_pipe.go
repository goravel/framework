package process

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

type RunningPipe struct {
	commands []*exec.Cmd
	steps    []*contractsprocess.Step

	interWriters []*io.PipeWriter

	stdOutputBuffers []*bytes.Buffer
	stdErrorBuffers  []*bytes.Buffer

	doneChan chan struct{}
	result   contractsprocess.Result
	resultMu sync.RWMutex
}

func NewRunningPipe(commands []*exec.Cmd, steps []*contractsprocess.Step, interWriters []*io.PipeWriter, stdout, stderr []*bytes.Buffer) *RunningPipe {
	run := &RunningPipe{
		commands:         commands,
		steps:            steps,
		interWriters:     interWriters,
		stdOutputBuffers: stdout,
		stdErrorBuffers:  stderr,
		doneChan:         make(chan struct{}),
	}

	go func() {
		var waitErr error
		for _, cmd := range run.commands {
			if cmd == nil {
				continue
			}
			waitErr = cmd.Wait()
		}

		lastIdx := len(run.commands) - 1
		var finalCmd *exec.Cmd
		if lastIdx >= 0 {
			finalCmd = run.commands[lastIdx]
		}

		exitCode := getExitCode(finalCmd, waitErr)

		cmdStr := ""
		if finalCmd != nil {
			cmdStr = finalCmd.String()
		}

		stdoutStr, stderrStr := "", ""
		if lastIdx >= 0 && lastIdx < len(run.stdOutputBuffers) && run.stdOutputBuffers[lastIdx] != nil {
			stdoutStr = run.stdOutputBuffers[lastIdx].String()
		}
		if lastIdx >= 0 && lastIdx < len(run.stdErrorBuffers) && run.stdErrorBuffers[lastIdx] != nil {
			stderrStr = run.stdErrorBuffers[lastIdx].String()
		}

		res := NewResult(exitCode, cmdStr, stdoutStr, stderrStr)

		run.resultMu.Lock()
		run.result = res
		run.resultMu.Unlock()

		for _, w := range run.interWriters {
			_ = w.Close()
		}

		close(run.doneChan)
	}()

	return run
}

func (r *RunningPipe) PIDs() map[string]int {
	m := make(map[string]int, len(r.commands))
	for i, cmd := range r.commands {
		key := ""
		if r.steps != nil && i < len(r.steps) && r.steps[i] != nil {
			key = r.steps[i].Key
		}
		if cmd != nil && cmd.Process != nil {
			m[key] = cmd.Process.Pid
		} else {
			m[key] = 0
		}
	}
	return m
}

func (r *RunningPipe) Running() bool {
	for _, cmd := range r.commands {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		if running(cmd.Process) {
			return true
		}
	}
	return false
}

func (r *RunningPipe) Done() <-chan struct{} {
	return r.doneChan
}

func (r *RunningPipe) Wait() contractsprocess.Result {
	<-r.doneChan
	r.resultMu.RLock()
	defer r.resultMu.RUnlock()
	return r.result
}

func (r *RunningPipe) Signal(sig os.Signal) error {
	var firstErr error
	for _, cmd := range r.commands {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		if running(cmd.Process) {
			if err := signal(cmd.Process, sig); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
	}
	return firstErr
}

func (r *RunningPipe) Stop(timeout time.Duration, sig ...os.Signal) error {
	var firstErr error
	for _, cmd := range r.commands {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		if err := stop(cmd.Process, r.doneChan, timeout, sig...); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
