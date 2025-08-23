package process

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/goravel/framework/contracts/process"
)

type Running struct {
	execCmd      *exec.Cmd
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
	startTime    time.Time

	waitOnce sync.Once
	result   *Result
}

func NewRunning(execCmd *exec.Cmd, stdout, stderr *bytes.Buffer) *Running {
	return &Running{
		execCmd:      execCmd,
		stdoutBuffer: stdout,
		stderrBuffer: stderr,
		startTime:    time.Now(),
	}
}

func (r *Running) Wait() process.Result {
	r.waitOnce.Do(func() {
		err := r.execCmd.Wait()
		r.result = buildResult(r, err)
	})
	return r.result
}

func (r *Running) PID() int {
	if r.execCmd.Process == nil {
		return 0
	}
	return r.execCmd.Process.Pid
}

func (r *Running) Command() string {
	return r.execCmd.String()
}

func (r *Running) Running() bool {
	state := r.execCmd.ProcessState
	return state == nil || !state.Success() || state.Exited()
}

func (r *Running) Kill() error {
	return r.Signal(syscall.SIGKILL)
}

func (r *Running) Signal(sig os.Signal) error {
	if r.execCmd.Process == nil {
		return errors.New("process not started")
	}
	return r.execCmd.Process.Signal(sig)
}

func (r *Running) Process() *os.Process {
	return r.execCmd.Process
}

func (r *Running) Output() string {
	return r.stdoutBuffer.String()
}

func (r *Running) ErrorOutput() string {
	return r.stderrBuffer.String()
}

func buildResult(r *Running, waitErr error) *Result {
	duration := time.Since(r.startTime)

	exitCode := 0
	if r.execCmd.ProcessState != nil {
		exitCode = r.execCmd.ProcessState.ExitCode()
	} else if waitErr != nil {
		exitCode = -1
	}

	return &Result{
		exitCode:     exitCode,
		command:      r.Command(),
		duration:     duration,
		processState: r.execCmd.ProcessState,
		stdout:       r.stdoutBuffer.String(),
		stderr:       r.stderrBuffer.String(),
	}
}
