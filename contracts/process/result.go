package process

// Result represents the outcome of a finished process execution.
// It provides access to exit status, captured output, and helper
// methods for inspecting process behavior.
type Result interface {
	// Successful reports whether the process exited with a zero exit code.
	Successful() bool

	// Failed reports whether the process exited with a non-zero exit code.
	Failed() bool

	// ExitCode returns the process exit code. A zero value typically
	// indicates success, while non-zero indicates failure.
	ExitCode() int

	// Output returns the full contents written to stdout by the process.
	Output() string

	// ErrorOutput returns the full contents written to stderr by the process.
	ErrorOutput() string

	// Error returns any error encountered during process execution (Go-related error).
	Error() error

	// Command returns the full command string used to start the process,
	// including program name and arguments.
	Command() string

	// SeeInOutput reports whether the given substring is present
	// in the process stdout output.
	SeeInOutput(needle string) bool

	// SeeInErrorOutput reports whether the given substring is present
	// in the process stderr output.
	SeeInErrorOutput(needle string) bool
}
