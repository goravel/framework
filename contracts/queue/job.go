package queue

import (
	"time"
)

type Job interface {
	// Signature set the unique signature of the job.
	Signature() string
	// Handle executes the job.
	Handle(args ...any) error
}

type PendingJob interface {
	// Delay dispatches the task after the given delay.
	Delay(time time.Time) PendingJob
	// Dispatch dispatches the task.
	Dispatch() error
	// DispatchSync dispatches the task synchronously.
	DispatchSync() error
	// OnConnection sets the connection of the task.
	OnConnection(connection string) PendingJob
	// OnQueue sets the single queue to which the task is dispatched.
	OnQueue(queue string) PendingJob
}

type ReservedJob interface {
	// Attempts returns the number of times the job has been attempted so far.
	Attempts() int
	// Delete removes the job from the queue.
	Delete() error
	// Release makes the job available again after the given delay so it can
	// be retried, incrementing attempts on the next pop.
	Release(delay time.Duration) error
	// Task returns the task to execute.
	Task() Task
}

type JobStorer interface {
	All() []Job
	Call(signature string, args []any) error
	Get(signature string) (Job, error)
	Register(jobs []Job)
}

// Deprecated: Use ChainJob instead.
type Jobs = ChainJob

type ChainJob struct {
	Delay time.Time `json:"delay"`
	Job   Job       `json:"job"`
	Args  []Arg     `json:"args"`
}

// JobWithUnique is implemented by jobs that must not be pushed again while
// one with the same unique ID is pending or running, like Laravel's
// ShouldBeUnique. Dispatch takes a cache lock for the job and skips it,
// returning nil, when the lock is already taken. The lock is released once
// the job succeeds or fails for good, not while it waits for a retry; the
// jobs chained to it don't hold it. DispatchSync does not check it. It needs
// a cache store shared by the dispatcher and the workers, such as Redis or
// the database: with a store local to each process, such as memory, every
// process takes its own lock.
type JobWithUnique interface {
	// UniqueID returns what tells apart the jobs with the same signature,
	// from the values the job is dispatched with, as Handle receives them.
	// An empty ID makes the job unique by its signature alone.
	UniqueID(args ...any) string
	// UniqueFor returns how long the lock is held at most, so a job that
	// never finishes does not block the next ones forever. Zero holds it
	// until the job is processed, without expiring.
	UniqueFor() time.Duration
}

type JobWithShouldRetry interface {
	// ShouldRetry determines if the job should be retried based on the error.
	// maxTries is the queue worker's configured tries: queue.Args.Tries for
	// Worker(args), or 1 for the no-argument Worker() (which hardcodes it).
	// Note that leaving Args.Tries unset passes the zero value 0; both 0 and 1
	// yield a single-shot fallback for jobs without their own retry policy.
	// Implementations without their own retry policy fall back to maxTries.
	ShouldRetry(err error, attempt, maxTries int) (retryable bool, delay time.Duration)
}
