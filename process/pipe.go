package process

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
	"github.com/goravel/framework/errors"
)

var _ contractsprocess.Pipe = (*Pipe)(nil)

func NewPipe() *Pipe {
	return &Pipe{
		ctx: context.Background(),
	}
}

type Pipe struct {
	ctx               context.Context
	input             io.Reader
	env               []string
	timeout           time.Duration
	onOutput          contractsprocess.OnPipeOutputFunc
	quietly           bool
	path              string
	bufferingDisabled bool
}

func (r *Pipe) DisableBuffering() contractsprocess.Pipe {
	r.bufferingDisabled = true
	return r
}

func (r *Pipe) Input(in io.Reader) contractsprocess.Pipe {
	r.input = in
	return r
}

func (r *Pipe) Env(vars map[string]string) contractsprocess.Pipe {
	if r.env == nil {
		r.env = make([]string, 0, len(vars))
	}
	for k, v := range vars {
		r.env = append(r.env, k+"="+v)
	}
	return r
}

func (r *Pipe) Path(path string) contractsprocess.Pipe {
	r.path = path
	return r
}

func (r *Pipe) Timeout(timeout time.Duration) contractsprocess.Pipe {
	r.timeout = timeout
	return r
}

func (r *Pipe) Quietly() contractsprocess.Pipe {
	r.quietly = true
	return r
}

func (r *Pipe) OnOutput(onOutput contractsprocess.OnPipeOutputFunc) contractsprocess.Pipe {
	r.onOutput = onOutput
	return r
}

func (r *Pipe) Run(configure func(contractsprocess.PipeBuilder)) (contractsprocess.Result, error) {
	return r.run(configure)
}

func (r *Pipe) Start(builder func(contractsprocess.PipeBuilder)) (contractsprocess.RunningPipe, error) {
	return r.start(builder)
}

func (r *Pipe) WithContext(ctx context.Context) contractsprocess.Pipe {
	if ctx == nil {
		ctx = context.Background()
	}

	r.ctx = ctx
	return r
}

func (r *Pipe) run(configure func(contractsprocess.PipeBuilder)) (contractsprocess.Result, error) {
	run, err := r.start(configure)
	if err != nil {
		return nil, err
	}
	return run.Wait(), nil
}

func (r *Pipe) start(configure func(contractsprocess.PipeBuilder)) (contractsprocess.RunningPipe, error) {
	builder := &PipeBuilder{}
	configure(builder)

	steps := builder.steps
	if len(steps) == 0 {
		return nil, errors.New("pipeline must have at least one command")
	}

	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc
	if r.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
	}

	commands := make([]*exec.Cmd, len(steps))
	for i, step := range steps {
		cmd := exec.CommandContext(ctx, step.Name, step.Args...)
		if r.path != "" {
			cmd.Dir = r.path
		}
		setSysProcAttr(cmd)

		if len(r.env) > 0 {
			cmd.Env = append(os.Environ(), r.env...)
		}

		commands[i] = cmd
	}

	// Prepare pipe connections between adjacent commands and configure stdout/stderr writers.
	// For i < len(commands)-1: command[i].Stdout -> pipeWriter -> pipeReader -> command[i+1].Stdin
	// Also, each command's stdout/stderr may also be copied to buffers, os.Stdout/os.Stderr and
	// an onOutput callback via MultiWriter.

	interReaders := make([]*io.PipeReader, 0, len(commands)-1)
	interWriters := make([]*io.PipeWriter, 0, len(commands)-1)

	stdoutBuffers := make([]*bytes.Buffer, len(commands))
	stderrBuffers := make([]*bytes.Buffer, len(commands))

	if len(commands) > 0 && r.input != nil {
		commands[0].Stdin = r.input
	}

	for i, cmd := range commands {
		var stdoutBuffer, stderrBuffer *bytes.Buffer
		var stdoutWriters []io.Writer
		var stderrWriters []io.Writer

		if !r.bufferingDisabled {
			stdoutBuffer = &bytes.Buffer{}
			stderrBuffer = &bytes.Buffer{}
			stdoutWriters = append(stdoutWriters, stdoutBuffer)
			stderrWriters = append(stderrWriters, stderrBuffer)
			stdoutBuffers[i] = stdoutBuffer
			stderrBuffers[i] = stderrBuffer
		}

		if !r.quietly {
			stdoutWriters = append(stdoutWriters, os.Stdout)
			stderrWriters = append(stderrWriters, os.Stderr)
		}

		if r.onOutput != nil {
			stdoutWriters = append(stdoutWriters, NewOutputWriterForPipe(steps[i].Key, contractsprocess.OutputTypeStdout, r.onOutput))
			stderrWriters = append(stderrWriters, NewOutputWriterForPipe(steps[i].Key, contractsprocess.OutputTypeStderr, r.onOutput))
		}

		// If this is not the last command, create a pipe to the next command and include the pipe writer
		// in this command's stdout MultiWriter — but ONLY if the next command does not already have stdin set.
		if i < len(commands)-1 {
			if commands[i+1].Stdin == nil {
				pr, pw := io.Pipe()
				interReaders = append(interReaders, pr)
				interWriters = append(interWriters, pw)
				stdoutWriters = append(stdoutWriters, pw)
				// set next command's stdin to the pipe reader
				commands[i+1].Stdin = pr
			}
		}

		if len(stdoutWriters) > 0 {
			cmd.Stdout = io.MultiWriter(stdoutWriters...)
		}

		if len(stderrWriters) > 0 {
			cmd.Stderr = io.MultiWriter(stderrWriters...)
		}
	}

	started := 0
	for i, cmd := range commands {
		if err := cmd.Start(); err != nil {
			if cancel != nil {
				cancel()
			}

			for j := 0; j < started; j++ {
				if commands[j].Process != nil {
					_ = kill(commands[j].Process)
				}
			}
			for _, w := range interWriters {
				_ = w.Close()
			}
			for _, r := range interReaders {
				_ = r.Close()
			}
			return nil, errors.New("failed to start pipeline: " + err.Error())
		}
		started = i + 1
	}

	return NewRunningPipe(commands, steps, cancel, interReaders, interWriters, stdoutBuffers, stderrBuffers), nil
}

type PipeBuilder struct {
	steps []*contractsprocess.Step
}

func (b *PipeBuilder) Command(name string, arg ...string) *contractsprocess.Step {
	step := &contractsprocess.Step{
		Key:  name,
		Name: name,
		Args: arg,
	}
	b.steps = append(b.steps, step)
	return step
}
