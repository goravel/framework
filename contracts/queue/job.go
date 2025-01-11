package queue

import "time"

type Job interface {
	// Signature set the unique signature of the job.
	Signature() string
	// Handle executes the job.
	Handle(args ...any) error
}

type JobRepository interface {
	All() []Job
	Call(signature string, args []any) error
	Get(signature string) (Job, error)
	Register(jobs []Job)
}

type Jobs struct {
	Job   Job
	Args  []any
	Delay time.Time
}
