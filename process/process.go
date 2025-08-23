package process

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.Process = (*Process)(nil)

type Process struct {
	args        []string
	env         []string
	forever     bool
	idleTimeout time.Duration
	input       io.Reader
	name        string
	path        string
	quietly     bool
	onOutput    contractsprocess.OnOutputFunc
	timeout     time.Duration
	tty         bool
}

func New() *Process {
	return &Process{}
}

func (r *Process) Command(name string, arg ...string) contractsprocess.Process {
	r.name = name
	r.args = append([]string(nil), arg...)
	return r
}

func (r *Process) Env(vars map[string]string) contractsprocess.Process {
	if r.env == nil {
		r.env = make([]string, 0, len(vars))
	}
	for k, v := range vars {
		r.env = append(r.env, k+"="+v)
	}
	return r
}

func (r *Process) Forever() contractsprocess.Process {
	r.forever = true
	return r
}

func (r *Process) IdleTimeout(timeout time.Duration) contractsprocess.Process {
	r.idleTimeout = timeout
	return r
}

func (r *Process) Input(in io.Reader) contractsprocess.Process {
	r.input = in
	return r
}

func (r *Process) Path(path string) contractsprocess.Process {
	r.path = path
	return r
}

func (r *Process) Quietly() contractsprocess.Process {
	r.quietly = true
	return r
}

func (r *Process) OnOutput(handler contractsprocess.OnOutputFunc) contractsprocess.Process {
	r.onOutput = handler
	return r
}

func (r *Process) Run(ctx context.Context) (contractsprocess.Result, error) {
	return r.run(ctx)
}

func (r *Process) Start(ctx context.Context) (contractsprocess.Running, error) {
	return r.start(ctx)
}

func (r *Process) Timeout(timeout time.Duration) contractsprocess.Process {
	r.timeout = timeout
	return r
}

func (r *Process) TTY() contractsprocess.Process {
	r.tty = true
	return r
}

func (r *Process) run(ctx context.Context) (contractsprocess.Result, error) {
	running, err := r.start(ctx)
	if err != nil {
		return nil, err
	}
	return running.Wait(), nil
}

func (r *Process) start(ctx context.Context) (contractsprocess.Running, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		_ = cancel
	}

	cmd := exec.CommandContext(ctx, r.name, r.args...)
	if r.path != "" {
		cmd.Dir = r.path
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if len(r.env) > 0 {
		cmd.Env = append(os.Environ(), r.env...)
	}

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	if r.input != nil {
		cmd.Stdin = r.input
	} else if r.tty {
		cmd.Stdin = os.Stdin
	}

	stdoutWriters := []io.Writer{stdoutBuffer}
	stderrWriters := []io.Writer{stderrBuffer}

	if !r.quietly {
		stdoutWriters = append(stdoutWriters, os.Stdout)
		stderrWriters = append(stderrWriters, os.Stderr)
		if r.onOutput != nil {
			stdoutWriters = append(stdoutWriters, NewOutputWriter(contractsprocess.OutputTypeStdout, r.onOutput))
			stderrWriters = append(stderrWriters, NewOutputWriter(contractsprocess.OutputTypeStderr, r.onOutput))
		}
	}
	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return NewRunning(cmd, stdoutBuffer, stderrBuffer), nil
}
