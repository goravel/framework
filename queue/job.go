package queue

import (
	"sync"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type FailedJob struct {
	ID        uint            `gorm:"primaryKey"`               // The unique ID of the job.
	Queue     string          `gorm:"not null"`                 // The name of the queue the job belongs to.
	Signature string          `gorm:"not null"`                 // The signature of the handler for this job.
	Payloads  []any           `gorm:"not null;serializer:json"` // The arguments passed to the job.
	Exception string          `gorm:"not null"`                 // The exception that caused the job to fail.
	FailedAt  carbon.DateTime `gorm:"not null"`                 // The timestamp when the job failed.
}

type Job interface {
	Register(jobs []contractsqueue.Job) error
	Call(signature string, args []any) error
	Get(signature string) (contractsqueue.Job, error)
	GetJobs() []contractsqueue.Job
}

type JobImpl struct {
	jobs sync.Map
}

func NewJobImpl() *JobImpl {
	return &JobImpl{}
}

// Register registers jobs to the job manager
func (r *JobImpl) Register(jobs []contractsqueue.Job) {
	for _, job := range jobs {
		r.jobs.Store(job.Signature(), job)
	}
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

	return nil, errors.New("job not found")
}

// GetJobs gets all registered jobs
func (r *JobImpl) GetJobs() []contractsqueue.Job {
	var jobs []contractsqueue.Job
	r.jobs.Range(func(_, value any) bool {
		jobs = append(jobs, value.(contractsqueue.Job))
		return true
	})

	return jobs
}
