package queue

type Job interface {
	// Signature set the unique signature of the job.
	Signature() string
	// Handle executes the job.
	Handle(args ...any) error
}

type Jobs struct {
	Job  Job
	Args []Arg
}
