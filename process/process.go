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

type Command struct {
	cmd *BaseCommand
}

func NewCommand(name string, args ...string) *Command {
	return &Command{
		cmd: &BaseCommand{
			name: name,
			args: args,
		},
	}
}

func (b *Command) Path(dir string) process.Command {
	b.cmd.Path(dir)
	return b
}

func (b *Command) Env(env map[string]string) process.Command {
	b.cmd.Env(env)
	return b
}

func (b *Command) Input(reader io.Reader) process.Command {
	b.cmd.Input(reader)
	return b
}

func (b *Command) Timeout(duration time.Duration) process.Command {
	b.cmd.Timeout(duration)
	return b
}

func (b *Command) IdleTimeout(duration time.Duration) process.Command {
	b.cmd.IdleTimeout(duration)
	return b
}

func (b *Command) Quietly() process.Command {
	b.cmd.Quietly()
	return b
}

func (b *Command) Tty() process.Command {
	b.cmd.Tty()
	return b
}

func (b *Command) OnOutput(handler func(typ, line string)) process.Command {
	b.cmd.OnOutput(handler)
	return b
}

func (b *Command) Run(ctx context.Context) (process.Result, error) {
	running, err := b.Start(ctx)
	if err != nil {
		return &Result{
			exitCode: -1,
			command:  b.cmd.name,
			stderr:   err.Error(),
		}, err
	}
	return running.Wait(), nil
}

func (b *Command) Start(ctx context.Context) (process.Running, error) {
	ctx, _ = b.prepareContext(ctx)

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

func (b *Command) prepareContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	if b.cmd.timeout > 0 {
		fmt.Println("timeout")
		return context.WithTimeout(ctx, b.cmd.timeout)
	}
	return context.WithCancel(ctx)
}

func (b *Command) configureCmd(execCmd *exec.Cmd) {
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
