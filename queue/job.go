package queue

import (
	"github.com/samber/do/v2"

	contractsqueue "github.com/goravel/framework/contracts/queue"
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
	injector   do.Injector
	signatures []string
}

func NewJobImpl() *JobImpl {
	return &JobImpl{
		injector: do.New(),
	}
}

// Register registers jobs to the injector
// Register 将作业注册到注入器
func (r *JobImpl) Register(jobs []contractsqueue.Job) error {
	for _, job := range jobs {
		do.ProvideNamedValue(r.injector, job.Signature(), job)
		r.signatures = append(r.signatures, job.Signature())
	}

	return nil
}

// Call calls a registered job using its signature
// Call 使用其签名调用已注册的作业
func (r *JobImpl) Call(signature string, args []any) error {
	job, err := do.InvokeNamed[contractsqueue.Job](r.injector, signature)
	if err != nil {
		return err
	}

	return job.Handle(args...)
}

// Get gets a registered job using its signature
// Get 使用其签名获取已注册的作业
func (r *JobImpl) Get(signature string) (contractsqueue.Job, error) {
	return do.InvokeNamed[contractsqueue.Job](r.injector, signature)
}

// GetJobs gets all registered jobs
// GetJobs 获取所有已注册的作业
func (r *JobImpl) GetJobs() []contractsqueue.Job {
	var jobs []contractsqueue.Job
	for _, signature := range r.signatures {
		job, err := r.Get(signature)
		if err != nil {
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs
}
