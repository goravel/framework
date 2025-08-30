package process

import (
	"context"
	"io"
	"time"
)

// OutputType represents the type of output stream produced by a running process.
type OutputType int

const (
	// OutputTypeStdout indicates output written to the standard output stream.
	OutputTypeStdout OutputType = iota

	// OutputTypeStderr indicates output written to the standard error stream.
	OutputTypeStderr
)

// OnOutputFunc is a callback function invoked when the process produces output.
// The typ parameter indicates whether the data came from stdout or stderr,
// and line contains the raw output bytes (typically a line of text).
type OnOutputFunc func(typ OutputType, line []byte)

// Process defines an interface for configuring and running external processes.
//
// Implementations are mutable and should not be reused concurrently.
// Each method modifies the same underlying process configuration.
type Process interface {
	// DisableBuffering prevents the process's stdout and stderr from being buffered
	// in memory. This is a critical optimization for commands that produce a large
	// volume of output, especially when that output is already being handled by
	// an OnOutput streaming callback.
	//
	// CONSEQUENCE: As output is not captured, the following methods on the
	// Running and Result handles will always return an empty string:
	//   - Running.Output()
	//   - Running.ErrorOutput()
	//   - Result.Output()
	//   - Result.ErrorOutput()
	DisableBuffering() Process

	// Env adds or overrides environment variables for the process.
	// Modifies the current process configuration.
	Env(vars map[string]string) Process

	// Input sets the stdin source for the process.
	// By default, processes run without stdin input.
	Input(in io.Reader) Process

	// Path sets the working directory where the process will be executed.
	Path(path string) Process

	// Quietly suppresses all process output, discarding both stdout and stderr.
	Quietly() Process

	// OnOutput registers a handler to receive stdout and stderr output
	// while the process runs. Multiple handlers may be chained depending
	// on the implementation.
	OnOutput(handler OnOutputFunc) Process

	// Run starts the process, waits for it to complete, and returns the result.
	// It returns an error if the process cannot be started or if execution fails.
	Run(name string, arg ...string) (Result, error)

	// Start begins running the process asynchronously and returns a Running
	// handle to monitor and control its execution. The caller must later
	// wait or terminate the process explicitly.
	Start(name string, arg ...string) (Running, error)

	// Timeout sets a maximum execution duration for the process.
	// If the timeout is exceeded, the process will be terminated.
	// A zero duration disables the timeout.
	Timeout(timeout time.Duration) Process

	// TTY attaches the process to a pseudo-terminal, enabling interactive
	// behavior (such as programs that require a TTY for input/output).
	TTY() Process

	// WithContext binds the process lifecycle to the provided context.
	// If the context is canceled or reaches its deadline, the process
	// will be terminated. When combined with Timeout, the earlier of
	// the two deadlines takes effect.
	WithContext(ctx context.Context) Process
}
