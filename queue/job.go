package queue

import (
	"github.com/samber/do/v2"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type FailedJob struct {
	ID        uint                 `gorm:"primaryKey"`               // The unique ID of the job.
	Queue     string               `gorm:"not null"`                 // The name of the queue the job belongs to.
	Signature string               `gorm:"not null"`                 // The signature of the handler for this job.
	Payloads  []contractsqueue.Arg `gorm:"not null;serializer:json"` // The arguments passed to the job.
	Exception string               `gorm:"not null"`                 // The exception that caused the job to fail.
	FailedAt  carbon.DateTime      `gorm:"not null"`                 // The timestamp when the job failed.
}

var injector = do.New()

// Register registers jobs to the registry.
// Register 将作业注册到注册表。
func Register(jobs []contractsqueue.Job) error {
	for _, job := range jobs {
		do.ProvideNamedValue(injector, job.Signature(), job)
	}

	return nil
}

// Call calls a registered job using its signature.
// Call 使用其签名调用已注册的作业。
func Call(signature string, args []contractsqueue.Arg) error {
	job, err := do.InvokeNamed[contractsqueue.Job](injector, signature)
	if err != nil {
		return err
	}

	return job.Handle(args)
}

// Get gets a registered job using its signature.
// Get 使用其签名获取已注册的作业。
func Get(signature string) (contractsqueue.Job, error) {
	return do.InvokeNamed[contractsqueue.Job](injector, signature)
}
