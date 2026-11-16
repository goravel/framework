package queue

import "context"

type Queue interface {
	// Connection gets a driver instance by connection name
	Connection(name string) (Driver, error)
	// Chain creates a chain of jobs to be processed one by one, passing
	Chain(jobs []ChainJob) PendingJob
	// Failer gets failed jobs
	Failer() Failer
	// GetJob gets job by signature
	GetJob(signature string) (Job, error)
	// GetJobs gets all jobs
	GetJobs() []Job
	// JobStorer gets job storer
	JobStorer() JobStorer
	// Job add a job to queue
	Job(job Job, args ...[]Arg) PendingJob
	// Register register jobs
	Register(jobs []Job)
	// Worker create a queue worker
	Worker(payloads ...Args) Worker
}

type Worker interface {
	// Process runs a reserved job received outside of the worker loop (e.g.
	// pushed by Lambda SQS event source mappings or Pub/Sub push
	// subscriptions) with the same semantics as Run: retries through Release,
	// chains, failed jobs and Delete. A failed job is recorded under the first
	// queue name of the worker's queue list.
	//
	// released reports that the job did not finish and must stay in the
	// queue: it was released for a retry (err is nil), or ctx was done before
	// the job or its chain finished (err is the context error; a partly run
	// chain is run again on redelivery). Otherwise the job is done: err is nil
	// on success, or the reason it failed for good. When deleting the job
	// fails, err also matches errors.QueueFailedToDeleteReservedJob and the
	// job is still in the queue.
	Process(ctx context.Context, reservedJob ReservedJob) (released bool, err error)
	Run() error
	Shutdown() error
}

type Args struct {
	// Specify connection
	Connection string
	// Queue specifies queue names as a comma-separated priority list,
	// consumed from left to right.
	Queue string
	// Concurrent num
	Concurrent int
	// Tries maximum attempts
	Tries int
}

type Arg struct {
	Value any    `json:"value"`
	Type  string `json:"type"`
}
