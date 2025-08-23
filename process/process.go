package process

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/goravel/framework/contracts/process"
)

func Run(ctx context.Context, name string, args ...string) (process.Result, error) {
	return Build(ctx, name, args...).Run()
}

func Start(ctx context.Context, name string, args ...string) (process.Running, error) {
	return Build(ctx, name, args...).Start()
}

func Build(ctx context.Context, name string, args ...string) process.SingleBuilder {
	cmd := NewCommand(ctx, name, args...)
	return &SingleBuilder{cmd: cmd}
}

type SingleBuilder struct {
	cmd *Command
}

func (b *SingleBuilder) Path(dir string) process.SingleBuilder {
	b.cmd.Path(dir)
	return b
}

func (b *SingleBuilder) Env(env map[string]string) process.SingleBuilder {
	b.cmd.Env(env)
	return b
}

func (b *SingleBuilder) Input(reader io.Reader) process.SingleBuilder {
	b.cmd.Input(reader)
	return b
}

func (b *SingleBuilder) Timeout(duration time.Duration) process.SingleBuilder {
	b.cmd.Timeout(duration)
	return b
}

func (b *SingleBuilder) IdleTimeout(duration time.Duration) process.SingleBuilder {
	b.cmd.IdleTimeout(duration)
	return b
}

func (b *SingleBuilder) Quietly() process.SingleBuilder {
	b.cmd.Quietly()
	return b
}

func (b *SingleBuilder) Tty() process.SingleBuilder {
	b.cmd.Tty()
	return b
}

func (b *SingleBuilder) OnOutput(handler func(typ, line string)) process.SingleBuilder {
	b.cmd.OnOutput(handler)
	return b
}

func (b *SingleBuilder) Run() (process.Result, error) {
	running, err := b.Start()
	if err != nil {
		return &Result{
			exitCode: -1,
			command:  b.cmd.name,
			stderr:   err.Error(),
		}, err
	}
	return running.Wait(), nil
}

func (b *SingleBuilder) Start() (process.Running, error) {
	ctx, _ := b.prepareContext()

	execCmd := exec.CommandContext(ctx, b.cmd.name, b.cmd.args...)
	b.configureCmd(execCmd)

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	if b.cmd.stdin != nil {
		execCmd.Stdin = b.cmd.stdin
	} else if b.cmd.tty {
		execCmd.Stdin = os.Stdin
	}

	stdoutWriters := []io.Writer{stdoutBuffer}
	stderrWriters := []io.Writer{stderrBuffer}
	if !b.cmd.quietly {
		stdoutWriters = append(stdoutWriters, os.Stdout)
		stderrWriters = append(stderrWriters, os.Stderr)
	}
	if b.cmd.outputHandler != nil {
		stdoutWriters = append(stdoutWriters, &OutputWriter{handler: b.cmd.outputHandler, typ: "stdout"})
		stderrWriters = append(stderrWriters, &OutputWriter{handler: b.cmd.outputHandler, typ: "stderr"})
	}
	execCmd.Stdout = io.MultiWriter(stdoutWriters...)
	execCmd.Stderr = io.MultiWriter(stderrWriters...)

	if err := execCmd.Start(); err != nil {
		return nil, err
	}

	if execCmd.Process != nil {
		syscall.Setpgid(execCmd.Process.Pid, execCmd.Process.Pid)
	}

	return NewRunning(execCmd, stdoutBuffer, stderrBuffer), nil
}

func (b *SingleBuilder) prepareContext() (context.Context, context.CancelFunc) {
	ctx := b.cmd.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if b.cmd.timeout > 0 {
		fmt.Println("timeout")
		return context.WithTimeout(ctx, b.cmd.timeout)
	}
	return context.WithCancel(ctx)
}

func (b *SingleBuilder) configureCmd(execCmd *exec.Cmd) {
	execCmd.Dir = b.cmd.dir
	if len(b.cmd.env) > 0 {
		execCmd.Env = append(os.Environ(), b.cmd.env...)
	}
	execCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

type OutputWriter struct {
	typ     string
	handler func(typ, line string)
}

func (w *OutputWriter) Write(p []byte) (n int, err error) {
	w.handler(w.typ, string(p))
	return len(p), nil
}
