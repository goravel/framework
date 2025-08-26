package process

import (
	"os"
	"time"
)

// Running represents a handle to a process that has been started and is still active.
// It provides methods for inspecting process state, retrieving output, and controlling
// its lifecycle.
type Running interface {
	// PID returns the operating system process ID.
	PID() int

	// Running reports whether the process still exists according to the OS.
	//
	// NOTE: This may return true for a "zombie" process (terminated but not
	// reaped). You must eventually call Wait() to reap the process and release
	// resources. A common pattern is to poll Running() in a goroutine and then
	// call Wait() in the main flow to ensure cleanup.
	Running() bool

	// Output returns the complete stdout captured from the process so far.
	Output() string

	// ErrorOutput returns the complete stderr captured from the process so far.
	ErrorOutput() string

	// LatestOutput returns the most recent chunk of stdout produced by the process.
	// The definition of "latest" is implementation dependent.
	LatestOutput() string

	// LatestErrorOutput returns the most recent chunk of stderr produced by the process.
	// The definition of "latest" is implementation dependent.
	LatestErrorOutput() string

	// Wait blocks until the process exits and returns its final Result.
	// This call is required to reap the process and release system resources.
	Wait() Result

	// Stop attempts to gracefully stop the process by sending the provided signal
	// (defaulting to SIGTERM). If the process does not exit within the given timeout,
	// it is forcibly killed (SIGKILL).
	Stop(timeout time.Duration, sig ...os.Signal) error

	// Signal sends the given signal to the process. Returns an error if the process
	// has not been started or no longer exists.
	Signal(sig os.Signal) error
}
