package queue

import (
	"errors"
	"sync"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

var mutex sync.RWMutex

// Register registers jobs to the registry.
// Register 将作业注册到注册表。
func Register(jobs []contractsqueue.Job) error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, job := range jobs {
		signature := job.Signature()
		if _, exists := JobRegistry[signature]; exists {
			return errors.New("Job with signature " + signature + " already exists")
		}
		JobRegistry[signature] = job
	}

	return nil
}

// Call calls a registered job using its signature.
// Call 使用其签名调用已注册的作业。
func Call(signature string, args ...any) error {
	mutex.RLock()
	defer mutex.RUnlock()

	job, exists := JobRegistry[signature]
	if !exists {
		return errors.New("job not found")
	}
	return job.Handle(args...)
}

type Job struct {
	ID            uint                 `gorm:"primaryKey"`
	Queue         string               `gorm:"type:text;not null"`
	Job           string               `gorm:"type:text;not null"`
	Arg           []contractsqueue.Arg `gorm:"type:json;not null;serializer:json"`
	Attempts      uint                 `gorm:"type:bigint;not null;default:0"`
	MaxTries      *uint                `gorm:"type:bigint;default:null;default:0"`
	MaxExceptions *uint                `gorm:"type:bigint;default:null;default:0"`
	Exception     *string              `gorm:"type:text;default:null"`
	Backoff       uint                 `gorm:"type:bigint;not null;default:0"`     // A number of seconds to wait before retrying the job.
	Timeout       *uint                `gorm:"type:bigint;default:null;default:0"` // The number of seconds the job can run.
	TimeoutAt     *carbon.DateTime     `gorm:"column:timeout_at"`
	ReservedAt    *carbon.DateTime     `gorm:"column:reserved_at"`
	AvailableAt   *carbon.DateTime     `gorm:"column:available_at"`
	CreatedAt     *carbon.DateTime     `gorm:"column:created_at"`
	FailedAt      *carbon.DateTime     `gorm:"column:failed_at"`
}

func (j Job) Signature() string {
	return j.Job
}

func (j Job) Handle(args ...any) error {
	return Call(j.Job, args...)
}
