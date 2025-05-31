package queue

type Queue interface {
	// Chain creates a chain of jobs to be processed one by one, passing
	Chain(jobs []ChainJob) PendingJob
	// GetJob gets job by signature
	GetJob(signature string) (Job, error)
	// GetJobs gets all jobs
	GetJobs() []Job
	// GetJobStorer gets job storer
	GetJobStorer() JobStorer
	// Job add a job to queue
	Job(job Job, args ...[]Arg) PendingJob
	// Register register jobs
	Register(jobs []Job)
	// Worker create a queue worker
	Worker(payloads ...Args) Worker
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

type Arg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}
