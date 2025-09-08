package process

import (
	"context"
	"io"
	"time"
)

type OnPipeOutputFunc func(key string, typ OutputType, line []byte)

// Pipeline defines a builder-style API for constructing and running a sequence
// of commands connected via OS pipes. Implementations are mutable and should
// not be used concurrently. Each configuration method returns the same
// Pipeline instance to allow fluent chaining. The Run/Start methods spawn the
// processes according to the provided builder.
type Pipeline interface {
	// DisableBuffering prevents capture of stdout/stderr into memory buffers.
	// When disabled, Result.Output and Result.ErrorOutput will be empty strings.
	DisableBuffering() Pipeline

	// Env adds or overrides environment variables for all steps.
	Env(vars map[string]string) Pipeline

	// Input sets the stdin source for the first step in the pipeline.
	Input(in io.Reader) Pipeline

	// Path sets the working directory for all steps.
	Path(path string) Pipeline

	// Quietly discards live stdout/stderr instead of mirroring to os.Stdout/err.
	Quietly() Pipeline

	// OnOutput registers a handler that receives line-delimited output produced
	// by each step while the pipeline runs.
	OnOutput(handler OnPipeOutputFunc) Pipeline

	// Run builds, executes, waits for completion, and returns the final Result.
	Run(func(builder Pipe)) (Result, error)

	// Start builds and starts execution asynchronously, returning a RunningPipe.
	Start(func(builder Pipe)) (RunningPipe, error)

	// Timeout sets a maximum duration for the entire pipeline execution.
	Timeout(timeout time.Duration) Pipeline

	// WithContext binds pipeline execution to the provided context.
	WithContext(ctx context.Context) Pipeline
}

type Pipe interface {
	Command(name string, arg ...string) Step
}

type Step interface {
	As(key string) Step
}
