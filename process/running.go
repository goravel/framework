package process

import (
	"bytes"
	"os/exec"
	"sync"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.Running = (*Running)(nil)

type Running struct {
	cmd          *exec.Cmd
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
	doneChan     chan struct{}

	result   contractsprocess.Result
	resultMu sync.RWMutex
}

func NewRunning(cmd *exec.Cmd, stdout, stderr *bytes.Buffer) *Running {
	running := &Running{
		cmd:          cmd,
		stdoutBuffer: stdout,
		stderrBuffer: stderr,
		doneChan:     make(chan struct{}),
	}

	go func() {
		waitErr := running.cmd.Wait()

		res := buildResult(running, waitErr)
		running.resultMu.Lock()
		running.result = res
		running.resultMu.Unlock()

		close(running.doneChan)
	}()

	return running
}

func (r *Running) Done() <-chan struct{} {
	return r.doneChan
}

func (r *Running) Wait() contractsprocess.Result {
	<-r.doneChan

	r.resultMu.RLock()
	defer r.resultMu.RUnlock()
	return r.result
}

func (r *Running) PID() int {
	if r.cmd == nil || r.cmd.Process == nil {
		return 0
	}
	return r.cmd.Process.Pid
}

func (r *Running) Command() string {
	if r.cmd == nil {
		return ""
	}
	return r.cmd.String()
}

func (r *Running) Output() string {
	if r.stdoutBuffer == nil {
		return ""
	}
	return r.stdoutBuffer.String()
}

func (r *Running) ErrorOutput() string {
	if r.stderrBuffer == nil {
		return ""
	}
	return r.stderrBuffer.String()
}

func (r *Running) LatestOutput() string {
	return lastN(r.stdoutBuffer, 4096)
}

func (r *Running) LatestErrorOutput() string {
	return lastN(r.stderrBuffer, 4096)
}

func buildResult(r *Running, waitErr error) *Result {
	exitCode := 0
	if r.cmd != nil && r.cmd.ProcessState != nil {
		exitCode = r.cmd.ProcessState.ExitCode()
	} else if waitErr != nil {
		exitCode = -1
	}

	command := ""
	if r.cmd != nil {
		command = r.Command()
	}

	stdout := ""
	if r.stdoutBuffer != nil {
		stdout = r.stdoutBuffer.String()
	}
	stderr := ""
	if r.stderrBuffer != nil {
		stderr = r.stderrBuffer.String()
	}

	return NewResult(exitCode, command, stdout, stderr)
}

func lastN(buf *bytes.Buffer, n int) string {
	if buf == nil {
		return ""
	}
	s := buf.String()
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
