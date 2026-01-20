package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.Running = (*Running)(nil)
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.CompleteColor{TrueColor: "#3D8C8D", ANSI256: "30", ANSI: "6"})
type Running struct {
	ctx            context.Context
	cmd            *exec.Cmd
	cancel         context.CancelFunc
	loading        bool
	loadingMessage string

	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer

	doneChan chan struct{}
	result   contractsprocess.Result
}

func NewRunning(ctx context.Context, cmd *exec.Cmd, cancel context.CancelFunc, stdout, stderr *bytes.Buffer, loading bool, loadingMessage string) *Running {
	runner := &Running{
		ctx:            ctx,
		cmd:            cmd,
		cancel:         cancel,
		loading:        loading,
		loadingMessage: loadingMessage,
		stdoutBuffer:   stdout,
		stderrBuffer:   stderr,
		doneChan:       make(chan struct{}),
	}

	go func(runner *Running) {
		defer func() {
			if err := recover(); err != nil {
				if runner.stderrBuffer != nil {
					_, _ = runner.stderrBuffer.WriteString("panic: ")
					_, _ = fmt.Fprint(runner.stderrBuffer, err)
					_, _ = runner.stderrBuffer.WriteString("\n")
				}
			}
			if runner.cancel != nil {
				runner.cancel()
			}
			close(runner.doneChan)
		}()

		waitErr := runner.cmd.Wait()
		exitCode := getExitCode(runner.cmd, waitErr)

		cmdStr := runner.cmd.String()

		stdoutStr, stderrStr := "", ""
		if runner.stdoutBuffer != nil {
			stdoutStr = runner.stdoutBuffer.String()
		}
		if runner.stderrBuffer != nil {
			stderrStr = runner.stderrBuffer.String()
		}

		runner.result = NewResult(waitErr, exitCode, cmdStr, stdoutStr, stderrStr)
	}(runner)

	return runner
}

func (r *Running) Command() string {
	return r.cmd.String()
}

func (r *Running) Done() <-chan struct{} {
	return r.doneChan
}

func (r *Running) Wait() contractsprocess.Result {
	if err := r.spinner(func() error {
		<-r.Done()

		return nil
	}); err != nil {
		return NewResult(err, 1, "", "", "")
	}

	return r.result
}

func (r *Running) PID() int {
	if r.cmd.Process == nil {
		return 0
	}
	return r.cmd.Process.Pid
}

func (r *Running) Running() bool {
	return running(r.cmd.Process)
}

func (r *Running) Kill() error {
	return kill(r.cmd.Process)
}

func (r *Running) Signal(sig os.Signal) error {
	return signal(r.cmd.Process, sig)
}

func (r *Running) Stop(timeout time.Duration, sig ...os.Signal) error {
	return stop(r.cmd.Process, r.doneChan, timeout, sig...)
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

func (r *Running) spinner(fn func() error) error {
	if !r.loading {
		return fn()
	}

	if r.loadingMessage == "" {
		r.loadingMessage = fmt.Sprintf("> %s", strings.Join(r.cmd.Args, " "))
	}

	spin := spinner.New().Title(r.loadingMessage).Style(spinnerStyle).TitleStyle(spinnerStyle)

	var err error
	spin = spin.Context(r.ctx).Action(func() {
		err = fn()
	})
	if err := spin.Run(); err != nil {
		return err
	}

	return err
}

func getExitCode(cmd *exec.Cmd, err error) int {
	exitCode := -1
	if cmd != nil && cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		}
	} else {
		// no error and no state -> assume 0
		exitCode = 0
	}

	return exitCode
}
