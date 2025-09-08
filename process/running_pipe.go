package process

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

type RunningPipe struct {
	commands []*exec.Cmd
	steps    []*Step
	cancel   context.CancelFunc

	interReaders []*io.PipeReader
	interWriters []*io.PipeWriter

	stdOutputBuffers []*bytes.Buffer
	stdErrorBuffers  []*bytes.Buffer

	doneChan chan struct{}
	result   contractsprocess.Result
}

func NewRunningPipe(
	commands []*exec.Cmd,
	steps []*Step,
	cancel context.CancelFunc,
	interReaders []*io.PipeReader,
	interWriters []*io.PipeWriter,
	stdout, stderr []*bytes.Buffer,
) *RunningPipe {
	pipeRunner := &RunningPipe{
		commands:         commands,
		steps:            steps,
		cancel:           cancel,
		interReaders:     interReaders,
		interWriters:     interWriters,
		stdOutputBuffers: stdout,
		stdErrorBuffers:  stderr,
		doneChan:         make(chan struct{}),
	}

	go func(runner *RunningPipe) {
		defer func() {
			if err := recover(); err != nil {
				// append panic to the last step's stderr buffer if available
				if len(runner.stdErrorBuffers) > 0 {
					lastIdx := len(runner.stdErrorBuffers) - 1
					if runner.stdErrorBuffers[lastIdx] != nil {
						runner.stdErrorBuffers[lastIdx].WriteString("panic: ")
						_, _ = fmt.Fprint(runner.stdErrorBuffers[lastIdx], err)
						runner.stdErrorBuffers[lastIdx].WriteString("\n")
					}
				}
			}
			if runner.cancel != nil {
				runner.cancel()
			}
			close(runner.doneChan)
		}()

		var waitErr error
		for i, cmd := range runner.commands {
			waitErr = cmd.Wait()

			// Close the writer that fed the next process's stdin.
			// Closing here ensures the next process sees EOF when upstream finishes.
			if i < len(runner.interWriters) {
				_ = runner.interWriters[i].Close()
			}
		}

		lastIdx := len(runner.commands) - 1
		finalCmd := runner.commands[lastIdx]

		exitCode := getExitCode(finalCmd, waitErr)

		cmdStr := finalCmd.String()

		stdoutStr, stderrStr := "", ""
		if runner.stdOutputBuffers[lastIdx] != nil {
			stdoutStr = runner.stdOutputBuffers[lastIdx].String()
		}
		if runner.stdErrorBuffers[lastIdx] != nil {
			stderrStr = runner.stdErrorBuffers[lastIdx].String()
		}

		runner.result = NewResult(waitErr, exitCode, cmdStr, stdoutStr, stderrStr)

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
		key := r.steps[i].key
		pid := 0
		if cmd.Process != nil {
			pid = cmd.Process.Pid
		}
		m[key] = pid
	}
	return m
}

func (r *RunningPipe) Running() bool {
	for _, cmd := range r.commands {
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
	<-r.Done()
	return r.result
}

func (r *RunningPipe) Signal(sig os.Signal) error {
	var firstErr error
	for _, cmd := range r.commands {
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
		if err := stop(cmd.Process, r.doneChan, timeout, sig...); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
