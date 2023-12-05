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
func Call(signature string, args []any) error {
	mutex.RLock()
	defer mutex.RUnlock()

	job, exists := JobRegistry[signature]
	if !exists {
		return errors.New("job not found")
	}

	return job.Handle(args...)
}

// Get gets a registered job using its signature.
// Get 使用其签名获取已注册的作业。
func Get(signature string) (contractsqueue.Job, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	job, exists := JobRegistry[signature]
	if !exists {
		return nil, errors.New("job not found")
	}

	return job, nil
}

type Job struct {
	ID            uint64           `gorm:"primaryKey" json:"id"`                        // The unique ID of the job.
	Queue         string           `gorm:"not null" json:"queue"`                       // The name of the queue the job belongs to.
	Job           string           `gorm:"not null" json:"job"`                         // The name of the handler for this job.
	Payloads      []any            `gorm:"not null;serializer:json" json:"payloads"`    // The arguments passed to the job.
	Attempts      uint             `gorm:"not null;default:0" json:"attempts"`          // The number of attempts made on the job.
	MaxTries      *uint            `gorm:"default:null;default:0" json:"maxTries"`      // The maximum number of attempts for this job.
	MaxExceptions *uint            `gorm:"default:null;default:0" json:"maxExceptions"` // The maximum number of exceptions to allow before failing.
	Backoff       uint             `gorm:"not null;default:0" json:"backoff"`           // The number of seconds to wait before retrying the job.
	Timeout       *uint            `gorm:"default:null;default:0" json:"timeout"`       // The number of seconds the job can run.
	TimeoutAt     *carbon.DateTime `gorm:"default:null" json:"timeoutAt"`               // The timestamp when the job running timeout.
	ReservedAt    *carbon.DateTime `gorm:"default:null" json:"reservedAt"`              // The timestamp when the job started running.
	AvailableAt   carbon.DateTime  `gorm:"not null" json:"availableAt"`                 // The timestamp when the job can start running.
	CreatedAt     carbon.DateTime  `gorm:"not null" json:"createdAt"`                   // The timestamp when the job was created.
}

func (j Job) Signature() string {
	return j.Job
}

func (j Job) Handle(args ...any) error {
	job, exists := JobRegistry[j.Signature()]
	if !exists {
		return errors.New("job not found")
	}

	return job.Handle(args...)
}

type FailedJob struct {
	ID        uint            `gorm:"primaryKey"`               // The unique ID of the job.
	Queue     string          `gorm:"not null"`                 // The name of the queue the job belongs to.
	Job       string          `gorm:"not null"`                 // The name of the handler for this job.
	Payloads  []any           `gorm:"not null;serializer:json"` // The arguments passed to the job.
	Exception string          `gorm:"not null"`                 // The exception that caused the job to fail.
	FailedAt  carbon.DateTime `gorm:"not null"`                 // The timestamp when the job failed.
}
