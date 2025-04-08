package queue

import (
	"sync"

	"github.com/google/uuid"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type FailedJob struct {
	ID         uint                 `gorm:"primaryKey" db:"id"`           // The unique ID of the job.
	UUID       uuid.UUID            `db:"uuid"`                           // The UUID of the job.
	Connection string               `db:"connection"`                     // The name of the connection the job belongs to.
	Queue      string               `db:"queue"`                          // The name of the queue the job belongs to.
	Payload    []contractsqueue.Arg `gorm:"serializer:json" db:"payload"` // The arguments passed to the job.
	Exception  string               `db:"exception"`                      // The exception that caused the job to fail.
	FailedAt   carbon.DateTime      `db:"failed_at"`                      // The timestamp when the job failed.
}

type JobRepository struct {
	jobs sync.Map
}

func NewJobRepository() *JobRepository {
	return &JobRepository{}
}

// All gets all registered jobs
func (r *JobRepository) All() []contractsqueue.Job {
	var jobs []contractsqueue.Job
	r.jobs.Range(func(_, value any) bool {
		jobs = append(jobs, value.(contractsqueue.Job))
		return true
	})

	return jobs
}

// Call calls a registered job using its signature
func (r *JobRepository) Call(signature string, args []any) error {
	job, err := r.Get(signature)
	if err != nil {
		return err
	}

	return job.Handle(args...)
}

// Get gets a registered job using its signature
func (r *JobRepository) Get(signature string) (contractsqueue.Job, error) {
	if job, ok := r.jobs.Load(signature); ok {
		return job.(contractsqueue.Job), nil
	}

	return nil, errors.QueueJobNotFound.Args(signature)
}

// Register registers jobs to the job manager
func (r *JobRepository) Register(jobs []contractsqueue.Job) {
	for _, job := range jobs {
		r.jobs.Store(job.Signature(), job)
	}
}
