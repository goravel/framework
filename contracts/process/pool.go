package process

import (
	"context"
	"io"
	"time"
)

type OnPoolOutputFunc func(key string, typ OutputType, line []byte)

// PoolBuilder defines the interface for configuring and launching a pool of concurrent processes.
type PoolBuilder interface {
	// WithConcurrency sets the maximum number of processes that can run simultaneously.
	// If n is zero or less, a default value (e.g., the number of tasks) will be used.
	WithConcurrency(n int) PoolBuilder

	// WithTimeout sets a total time limit for the entire pool operation. If the
	// timeout is exceeded, all running processes will be terminated.
	WithTimeout(timeout time.Duration) PoolBuilder

	// WithContext binds the pool's lifecycle to the provided context. If the context
	// is canceled, all running processes will be terminated.
	WithContext(ctx context.Context) PoolBuilder

	// WithOutputHandler sets a handler to receive real-time output from all
	// processes in the pool.
	WithOutputHandler(handler OnPoolOutputFunc) PoolBuilder

	// WithStrategy sets a custom scheduling algorithm for the pool. The default
	// strategy is FIFO (First-In, First-Out).
	WithStrategy(strategy Strategy) PoolBuilder

	// Run starts the pool, waits for all processes to complete, and returns a
	// map of the results, keyed by the process keys.
	Run(builder func(Pool)) (map[string]Result, error)

	// Start launches the pool asynchronously and returns a handle to the running
	// pool, allowing for interaction with the live processes.
	Start(builder func(Pool)) (RunningPool, error)
}

type Pool interface {
	Command(name string, arg ...string) PoolCommand
}

// PoolCommand provides a builder interface for a single command within a pool.
// It must satisfy the Schedulable interface to be used by scheduling strategies.
type PoolCommand interface {
	Schedulable

	// WithKey assigns a unique string key to the process. This key is used
	// to identify the process in the final results map and in the output handler.
	WithKey(key string) PoolCommand

	// WithContext binds the lifecycle of this specific process to the provided
	// context, overriding the pool's context if set.
	WithContext(ctx context.Context) PoolCommand

	// WithTimeout sets a maximum execution duration for this specific process,
	// overriding the pool's timeout if set.
	WithTimeout(timeout time.Duration) PoolCommand

	// WithPath sets the working directory for this specific process.
	WithPath(path string) PoolCommand

	// WithEnv sets environment variables for this specific process.
	WithEnv(vars map[string]string) PoolCommand

	// WithQuiet suppresses all process output from this specific process,
	// preventing it from being captured or sent to the output handler.
	WithQuiet() PoolCommand

	// WithDisabledBuffering prevents this process's output from being buffered in memory.
	// This is a critical optimization for commands with large output volumes.
	// Note: The result for this process will have an empty output.
	WithDisabledBuffering() PoolCommand

	// WithInput sets the stdin source for this specific process.
	WithInput(in io.Reader) PoolCommand

	// WithPriority priority
	WithPriority(priority Priority) PoolCommand
}
