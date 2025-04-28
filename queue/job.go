package queue

import (
	"sync"

	"github.com/google/uuid"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type FailedJob struct {
	ID         uint             `gorm:"primaryKey"` // The unique ID of the job.
	UUID       uuid.UUID        // The UUID of the job.
	Connection string           // The name of the connection the job belongs to.
	Queue      string           // The name of the queue the job belongs to.
	Payload    []any            `gorm:"serializer:json"` // The arguments passed to the job.
	Exception  string           // The exception that caused the job to fail.
	FailedAt   *carbon.DateTime // The timestamp when the job failed.
}

type JobImpl struct {
	jobs sync.Map
}

func NewJobImpl() *JobImpl {
	return &JobImpl{}
}

// All gets all registered jobs
func (r *JobImpl) All() []contractsqueue.Job {
	var jobs []contractsqueue.Job
	r.jobs.Range(func(_, value any) bool {
		jobs = append(jobs, value.(contractsqueue.Job))
		return true
	})

	return jobs
}

// Call calls a registered job using its signature
func (r *JobImpl) Call(signature string, args []any) error {
	job, err := r.Get(signature)
	if err != nil {
		return err
	}

	return job.Handle(args...)
}

// Get gets a registered job using its signature
func (r *JobImpl) Get(signature string) (contractsqueue.Job, error) {
	if job, ok := r.jobs.Load(signature); ok {
		return job.(contractsqueue.Job), nil
	}

	return nil, errors.QueueJobNotFound.Args(signature)
}

// Register registers jobs to the job manager
func (r *JobImpl) Register(jobs []contractsqueue.Job) {
	for _, job := range jobs {
		r.jobs.Store(job.Signature(), job)
	}
}
