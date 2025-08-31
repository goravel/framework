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

	interReaders []*io.PipeReader
	interWriters []*io.PipeWriter

	stdOutputBuffers []*bytes.Buffer
	stdErrorBuffers  []*bytes.Buffer

	doneChan chan struct{}
	result   contractsprocess.Result
	resultMu sync.RWMutex
}

func NewRunningPipe(
	commands []*exec.Cmd,
	steps []*contractsprocess.Step,
	interReaders []*io.PipeReader,
	interWriters []*io.PipeWriter,
	stdout, stderr []*bytes.Buffer,
) *RunningPipe {
	pipeRunner := &RunningPipe{
		commands:         commands,
		steps:            steps,
		interReaders:     interReaders,
		interWriters:     interWriters,
		stdOutputBuffers: stdout,
		stdErrorBuffers:  stderr,
		doneChan:         make(chan struct{}),
	}

	go func(runner *RunningPipe) {
		defer func() {
			if err := recover(); err != nil {
				// TODO: see what should be done here with this error, should we consider it as WaitErr?
			}
			close(runner.doneChan)
		}()

		var waitErr error
		for i, cmd := range runner.commands {
			if cmd == nil {
				// If there's an interWriter for this position, close it so downstream doesn't block
				if i < len(runner.interWriters) {
					_ = runner.interWriters[i].Close()
				}
				continue
			}

			waitErr = cmd.Wait()

			// Close the writer that fed the next process's stdin.
			// Closing here ensures the next process sees EOF when upstream finishes.
			if i < len(runner.interWriters) {
				_ = runner.interWriters[i].Close()
			}
		}

		lastIdx := len(runner.commands) - 1
		var finalCmd *exec.Cmd
		if lastIdx >= 0 {
			finalCmd = runner.commands[lastIdx]
		}

		exitCode := getExitCode(finalCmd, waitErr)

		cmdStr := ""
		if finalCmd != nil {
			cmdStr = finalCmd.String()
		}

		stdoutStr, stderrStr := "", ""
		if lastIdx >= 0 && lastIdx < len(runner.stdOutputBuffers) && runner.stdOutputBuffers[lastIdx] != nil {
			stdoutStr = runner.stdOutputBuffers[lastIdx].String()
		}
		if lastIdx >= 0 && lastIdx < len(runner.stdErrorBuffers) && runner.stdErrorBuffers[lastIdx] != nil {
			stderrStr = runner.stdErrorBuffers[lastIdx].String()
		}

		res := NewResult(exitCode, cmdStr, stdoutStr, stderrStr)

		runner.resultMu.Lock()
		runner.result = res
		runner.resultMu.Unlock()

		for _, w := range runner.interWriters {
			_ = w.Close()
		}
		for _, r := range runner.interReaders {
			_ = r.Close()
		}
	}(pipeRunner)

	return pipeRunner
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
