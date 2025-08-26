package process

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.Process = (*Process)(nil)

type Process struct {
	ctx      context.Context
	env      []string
	input    io.Reader
	path     string
	quietly  bool
	onOutput contractsprocess.OnOutputFunc
	timeout  time.Duration
	tty      bool
}

func New() *Process {
	return &Process{
		ctx: context.Background(),
	}
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

func (r *Process) Run(name string, args ...string) (contractsprocess.Result, error) {
	return r.run(name, args...)
}

func (r *Process) Start(name string, args ...string) (contractsprocess.Running, error) {
	return r.start(name, args...)
}

func (r *Process) Timeout(timeout time.Duration) contractsprocess.Process {
	r.timeout = timeout
	return r
}

func (r *Process) TTY() contractsprocess.Process {
	r.tty = true
	return r
}

func (r *Process) WithContext(ctx context.Context) contractsprocess.Process {
	if ctx == nil {
		ctx = context.Background()
	}

	r.ctx = ctx
	return r
}

func (r *Process) run(name string, args ...string) (contractsprocess.Result, error) {
	running, err := r.start(name, args...)
	if err != nil {
		return nil, err
	}
	return running.Wait(), nil
}

func (r *Process) start(name string, args ...string) (contractsprocess.Running, error) {
	ctx := r.ctx
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		_ = cancel
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if r.path != "" {
		cmd.Dir = r.path
	}
	setSysProcAttr(cmd)

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
	}
	if r.onOutput != nil {
		stdoutWriters = append(stdoutWriters, NewOutputWriter(contractsprocess.OutputTypeStdout, r.onOutput))
		stderrWriters = append(stderrWriters, NewOutputWriter(contractsprocess.OutputTypeStderr, r.onOutput))
	}
	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return NewRunning(cmd, stdoutBuffer, stderrBuffer), nil
}
