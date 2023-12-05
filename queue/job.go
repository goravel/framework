package queue

import (
	"errors"
	"reflect"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

// Register registers jobs to the registry.
// Register 将作业注册到注册表。
func Register(jobs []contractsqueue.Job) error {
	for _, job := range jobs {
		signature := job.Signature()
		if err := ValidateTask(job.Handle); err != nil {
			return err
		}
		if _, exists := JobRegistry.Load(signature); exists {
			return errors.New("Job with signature " + signature + " already exists")
		}

		JobRegistry.Store(signature, job)
	}

	return nil
}

// Call calls a registered job using its signature.
// Call 使用其签名调用已注册的作业。
func Call(signature string, args []contractsqueue.Arg) error {
	var err error
	defer func() {
		// Recover from panic and set err.
		if e := recover(); e != nil {
			switch e := e.(type) {
			default:
				err = errors.New("invoking handle caused a panic")
			case error:
				err = e
			case string:
				err = errors.New(e)
			}

			// TODO log the error
		}
	}()

	value, exists := JobRegistry.Load(signature)
	if !exists {
		return errors.New("job not found")
	}
	job, ok := value.(contractsqueue.Job)
	if !ok {
		return errors.New("job must implement contracts/queue/Job interface")
	}

	values, err := argsToValues(args)
	if err != nil {
		return err
	}

	// Invoke the handle
	results := reflect.ValueOf(job.Handle).Call(values)

	// Handle must return at least a value
	if len(results) == 0 {
		return ErrTaskReturnsNoValue
	}

	// Last returned value
	lastResult := results[len(results)-1]

	// If the last returned value is not nil, it has to be of error type, if that
	// is not the case, return error message, otherwise propagate the handle error
	// to the caller
	if !lastResult.IsNil() {
		// check that the result implements the standard error interface,
		// if not, return ErrLastReturnValueMustBeError error
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !lastResult.Type().Implements(errorInterface) {
			return ErrLastReturnValueMustBeError
		}

		// Return the standard error
		return lastResult.Interface().(error)
	}

	return nil
}

// Get gets a registered job using its signature.
// Get 使用其签名获取已注册的作业。
func Get(signature string) (contractsqueue.Job, error) {
	value, exists := JobRegistry.Load(signature)
	if !exists {
		return nil, errors.New("job not found")
	}
	job, ok := value.(contractsqueue.Job)
	if !ok {
		return nil, errors.New("job must implement contracts/queue/Job interface")
	}

	return job, nil
}

type Job struct {
	ID            uint64               `gorm:"primaryKey" json:"id"`                        // The unique ID of the job.
	Queue         string               `gorm:"not null" json:"queue"`                       // The name of the queue the job belongs to.
	Job           string               `gorm:"not null" json:"job"`                         // The name of the handler for this job.
	Payloads      []contractsqueue.Arg `gorm:"not null;serializer:json" json:"payloads"`    // The arguments passed to the job.
	Attempts      uint                 `gorm:"not null;default:0" json:"attempts"`          // The number of attempts made on the job.
	MaxTries      *uint                `gorm:"default:null;default:0" json:"maxTries"`      // The maximum number of attempts for this job.
	MaxExceptions *uint                `gorm:"default:null;default:0" json:"maxExceptions"` // The maximum number of exceptions to allow before failing.
	Backoff       uint                 `gorm:"not null;default:0" json:"backoff"`           // The number of seconds to wait before retrying the job.
	Timeout       *uint                `gorm:"default:null;default:0" json:"timeout"`       // The number of seconds the job can run.
	TimeoutAt     *carbon.DateTime     `gorm:"default:null" json:"timeoutAt"`               // The timestamp when the job running timeout.
	ReservedAt    *carbon.DateTime     `gorm:"default:null" json:"reservedAt"`              // The timestamp when the job started running.
	AvailableAt   carbon.DateTime      `gorm:"not null" json:"availableAt"`                 // The timestamp when the job can start running.
	CreatedAt     carbon.DateTime      `gorm:"not null" json:"createdAt"`                   // The timestamp when the job was created.
}

func (j Job) Signature() string {
	return j.Job
}

func (j Job) Handle(args ...any) error {
	job, err := Get(j.Signature())
	if err != nil {
		return err
	}

	return job.Handle(args...)
}

type FailedJob struct {
	ID        uint                 `gorm:"primaryKey"`               // The unique ID of the job.
	Queue     string               `gorm:"not null"`                 // The name of the queue the job belongs to.
	Job       string               `gorm:"not null"`                 // The name of the handler for this job.
	Payloads  []contractsqueue.Arg `gorm:"not null;serializer:json"` // The arguments passed to the job.
	Exception string               `gorm:"not null"`                 // The exception that caused the job to fail.
	FailedAt  carbon.DateTime      `gorm:"not null"`                 // The timestamp when the job failed.
}
