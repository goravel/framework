package queue

type Queue interface {
	Worker(payloads ...Args) Worker
	// Register register jobs
	Register(jobs []Job)
	// GetJobs get all jobs
	GetJobs() []Job
	// GetJob get job by signature
	GetJob(signature string) (Job, error)
	// Job add a job to queue
	Job(job Job, args []any) Task
	// Chain creates a chain of jobs to be processed one by one, passing
	Chain(jobs []Jobs) Task
}

type Worker interface {
	Run() error
	Shutdown() error
}

type Args struct {
	// Specify connection
	Connection string
	// Specify queue
	Queue string
	// Concurrent num
	Concurrent int
}
