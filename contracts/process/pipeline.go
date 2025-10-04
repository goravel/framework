package process

import (
	"context"
	"io"
	"time"
)

type OnPipeOutputFunc func(key string, typ OutputType, line []byte)

// Pipeline defines a builder-style API for constructing and running a sequence
// of commands connected via OS pipes.
//
// Implementations are mutable and should not be used concurrently. Each
// configuration method returns the same Pipeline instance to allow fluent chaining.
type Pipeline interface {
	// WithDisabledBuffering prevents capture of the final command's stdout/stderr
	// into memory buffers. When disabled, Result.Output and Result.ErrorOutput
	// will be empty strings.
	WithDisabledBuffering() Pipeline

	// WithEnv adds or overrides environment variables for all steps in the pipeline.
	// These can be overridden on a per-step basis via PipeCommand.
	WithEnv(vars map[string]string) Pipeline

	// WithInput sets the stdin source for the first step in the pipeline.
	WithInput(in io.Reader) Pipeline

	// WithPath sets the working directory for all steps in the pipeline.
	// This can be overridden on a per-step basis via PipeCommand.
	WithPath(path string) Pipeline

	// WithQuiet discards the live stdout/stderr of the final command instead
	// of mirroring it to the parent process's os.Stdout/err.
	WithQuiet() Pipeline

	// WithOutputHandler registers a handler that receives line-delimited output
	// produced by the final command while the pipeline runs.
	WithOutputHandler(handler OnPipeOutputFunc) Pipeline

	// WithTimeout sets a maximum duration for the entire pipeline execution.
	// If the timeout is exceeded, all processes in the pipeline will be terminated.
	WithTimeout(timeout time.Duration) Pipeline

	// WithContext binds the entire pipeline's lifecycle to the provided context.
	// If the context is canceled, all processes will be terminated.
	WithContext(ctx context.Context) Pipeline

	// Run builds, executes the pipeline, waits for completion, and returns the final Result.
	// The result is determined by the status of the final command in the pipe.
	Run(builder func(pipe Pipe)) (Result, error)

	// Start builds and starts the pipeline asynchronously, returning a RunningPipe handle.
	Start(builder func(pipe Pipe)) (RunningPipe, error)
}

type Pipe interface {
	Command(name string, arg ...string) PipeCommand
}

type PipeCommand interface {
	WithKey(key string) PipeCommand
}
