package process

import (
	"context"
	"io"
	"time"
)

type OnPoolOutputFunc func(typ OutputType, line []byte, key string)

// PoolBuilder defines the interface for configuring and launching a pool of concurrent processes.
type PoolBuilder interface {
	// Concurrency sets the maximum number of processes that can run simultaneously.
	// If n is zero or less, a default value (e.g., the number of tasks) will be used.
	Concurrency(n int) PoolBuilder

	// OnOutput sets a handler to receive real-time output from all
	// processes in the pool.
	OnOutput(handler OnPoolOutputFunc) PoolBuilder

	// Run starts the pool, waits for all processes to complete, and returns a
	// map of the results, keyed by the process keys.
	Run(builder func(Pool)) (map[string]Result, error)

	// Start launches the pool asynchronously and returns a handle to the running
	// pool, allowing for interaction with the live processes.
	Start(builder func(Pool)) (RunningPool, error)

	// Timeout sets a total time limit for the entire pool operation. If the
	// timeout is exceeded, all running processes will be terminated.
	Timeout(timeout time.Duration) PoolBuilder

	// WithContext binds the pool's lifecycle to the provided context. If the context
	// is canceled, all running processes will be terminated.
	WithContext(ctx context.Context) PoolBuilder
}

type Pool interface {
	Command(name string, arg ...string) PoolCommand
}

// PoolCommand provides a builder interface for a single command within a pool.
// It must satisfy the Schedulable interface to be used by scheduling strategies.
type PoolCommand interface {
	// As assigns a unique string key to the process. This key is used
	// to identify the process in the final results map and in the output handler.
	As(key string) PoolCommand

	// DisableBuffering prevents this process's output from being buffered in memory.
	// This is a critical optimization for commands with large output volumes.
	// Note: The result for this process will have an empty output.
	DisableBuffering() PoolCommand

	// Env sets environment variables for this specific process.
	Env(vars map[string]string) PoolCommand

	// Input sets the stdin source for this specific process.
	Input(in io.Reader) PoolCommand

	// Path sets the working directory for this specific process.
	Path(path string) PoolCommand

	// Quietly suppresses all process output from this specific process,
	// preventing it from being captured or sent to the output handler.
	Quietly() PoolCommand

	// Timeout sets a maximum execution duration for this specific process,
	// overriding the pool's timeout if set.
	Timeout(timeout time.Duration) PoolCommand

	// WithContext binds the lifecycle of this specific process to the provided
	// context, overriding the pool's context if set.
	WithContext(ctx context.Context) PoolCommand
}
