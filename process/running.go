package process

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.Running = (*Running)(nil)

type Running struct {
	cmd          *exec.Cmd
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer

	once     sync.Once
	result   contractsprocess.Result
	resultMu sync.RWMutex
}

func NewRunning(cmd *exec.Cmd, stdout, stderr *bytes.Buffer) *Running {
	return &Running{
		cmd:          cmd,
		stdoutBuffer: stdout,
		stderrBuffer: stderr,
	}
}

func (r *Running) Wait() contractsprocess.Result {
	r.once.Do(func() {
		err := r.cmd.Wait()
		res := buildResult(r, err)
		r.resultMu.Lock()
		r.result = res
		r.resultMu.Unlock()
	})
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

func (r *Running) Running() bool {
	if r.cmd == nil || r.cmd.Process == nil {
		return false
	}
	if r.cmd.ProcessState == nil {
		return true
	}
	return !r.cmd.ProcessState.Exited()
}

func (r *Running) Kill() error {
	return r.Signal(syscall.SIGKILL)
}

func (r *Running) Signal(sig os.Signal) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return errors.New("process not started")
	}
	return r.cmd.Process.Signal(sig)
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

func (r *Running) Stop(timeout time.Duration, sig ...os.Signal) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	signalToSend := syscall.SIGTERM
	if len(sig) > 0 {
		signalToSend = sig[0].(syscall.Signal)
	}

	if err := r.Signal(signalToSend); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return err
	}

	done := make(chan struct{})
	go func() {
		r.Wait()
		close(done)
	}()

	select {
	case <-time.After(timeout):
		return r.Signal(syscall.SIGKILL)
	case <-done:
		return nil
	}
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
