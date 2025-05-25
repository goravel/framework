package queue

import (
	"time"

	"github.com/goravel/framework/support/carbon"
)

type Job interface {
	// Signature set the unique signature of the job.
	Signature() string
	// Handle executes the job.
	Handle(args ...any) error
}

type JobRecord interface {
	Increment() int
	Touch() *carbon.DateTime
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
	// OnQueue sets the queue of the task.
	OnQueue(queue string) PendingJob
}

type ReservedJob interface {
	Delete() error
	Task() Task
}

type ReservedJobCreator interface {
	New(JobRecord) (ReservedJob, error)
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
	Job   Job       `json:"job"`
	Args  []Arg     `json:"args"`
	Delay time.Time `json:"delay"`
}
