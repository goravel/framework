package process

import (
	"os"
	"time"
)

// Running represents a handle to a single, active process.
// Its primary role is to manage the lifecycle and inspect the state of the process.
type Running interface {
	// PID returns the operating system process ID.
	PID() int

	// Running reports whether the process still exists according to the OS.
	// NOTE: This may return true for a "zombie" process (one that has terminated
	// but has not been reaped). Wait() must be called to reap the process.
	Running() bool

	// Done returns a read-only channel that is closed once the process has exited.
	//
	// This provides an efficient, non-polling mechanism to wait for process completion,
	// typically in a select statement.
	//
	// After receiving a signal from this channel, the caller is still required to
	// call Wait() to retrieve the process's final Result and to release all
	// underlying system resources. Failure to do so will result in a resource leak.
	Done() <-chan struct{}

	// Output returns the complete stdout captured from the process so far.
	//
	// WARNING: This method buffers the entire output in memory. For processes that
	// may generate a large volume of output, use the Process.OnOutput() callback during
	// configuration to stream the data instead.
	Output() string

	// ErrorOutput returns the complete stderr captured from the process so far.
	//
	// WARNING: This method buffers the entire output in memory. For processes that
	// may generate a large volume of output, use the Process.OnOutput() callback during
	// configuration to stream the data instead.
	ErrorOutput() string

	// LatestOutput returns the most recent chunk of stdout produced by the process.
	// TODO: Remove in a subsequent PR. The OnOutput callback is the superior, idiomatic
	// pattern for handling live-streaming output, making this method redundant.
	LatestOutput() string

	// LatestErrorOutput returns the most recent chunk of stderr produced by the process.
	// TODO: Remove in a subsequent PR. The OnOutput callback is the superior, idiomatic
	// pattern for handling live-streaming error output, making this method redundant.
	LatestErrorOutput() string

	// Wait blocks until the process exits and returns its final Result.
	// This call is required to reap the process and release all system resources.
	Wait() Result

	// Stop attempts to gracefully stop the process by sending the provided signal
	// (defaulting to SIGTERM). If the process does not exit within the given
	// timeout, it is forcibly killed (SIGKILL).
	Stop(timeout time.Duration, sig ...os.Signal) error

	// Signal sends the given signal to the process.
	Signal(sig os.Signal) error
}
