package queue

import (
	"github.com/goravel/framework/contracts/database/orm"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type Job struct {
	ID            uint64               `gorm:"primaryKey" json:"id"`                        // The unique ID of the job.
	Queue         string               `gorm:"not null" json:"queue"`                       // The name of the queue the job belongs to.
	Signature     string               `gorm:"not null" json:"signature"`                   // The signature of the handler for this job.
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

type FailedJob struct {
	ID        uint                 `gorm:"primaryKey"`               // The unique ID of the job.
	Queue     string               `gorm:"not null"`                 // The name of the queue the job belongs to.
	Signature string               `gorm:"not null"`                 // The signature of the handler for this job.
	Payloads  []contractsqueue.Arg `gorm:"not null;serializer:json"` // The arguments passed to the job.
	Exception string               `gorm:"not null"`                 // The exception that caused the job to fail.
	FailedAt  carbon.DateTime      `gorm:"not null"`                 // The timestamp when the job failed.
}

type Database struct {
	connection string
	query      orm.Query
}

func NewDatabase(connection string, query orm.Query) *Database {
	return &Database{
		connection: connection,
		query:      query,
	}
}

func (r *Database) Connection() string {
	return r.connection
}

func (r *Database) Driver() string {
	return DriverDatabase
}

func (r *Database) Push(job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	now := carbon.Now()
	var j Job
	j.Queue = queue
	j.Signature = job.Signature()
	j.Payloads = args
	j.AvailableAt = carbon.DateTime{Carbon: now}
	j.CreatedAt = carbon.DateTime{Carbon: now}

	return r.query.Create(&j)
}

func (r *Database) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	now := carbon.Now()
	var j []Job
	for _, job := range jobs {
		var jj Job
		jj.Queue = queue
		jj.Signature = job.Job.Signature()
		jj.Payloads = job.Args
		jj.AvailableAt = carbon.DateTime{Carbon: now.AddSeconds(int(job.Delay))}
		jj.CreatedAt = carbon.DateTime{Carbon: now}
		j = append(j, jj)
	}

	return r.query.Create(&j)
}

func (r *Database) Later(delay uint, job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	now := carbon.Now()
	var j Job
	j.Queue = queue
	j.Signature = job.Signature()
	j.Payloads = args
	j.AvailableAt = carbon.DateTime{Carbon: now.AddSeconds(int(delay))}
	j.CreatedAt = carbon.DateTime{Carbon: now}

	return r.query.Create(&j)
}

func (r *Database) Pop(q string) (contractsqueue.Job, []contractsqueue.Arg, error) {
	var job Job
	err := r.query.Where("queue", q).Where("reserved_at", nil).Where("available_at", "<=", carbon.DateTime{Carbon: carbon.Now()}).Order("id asc").First(&job)
	if err != nil {
		return nil, nil, err
	}
	handler, err := Get(job.Signature)

	return handler, job.Payloads, err
}

func (r *Database) Delete(queue string, job contractsqueue.Jobs) error {
	var j Job
	err := r.query.Where("queue", queue).Where("job", job.Job.Signature()).Where("payloads", job.Args).Order("id desc").FirstOrFail(&j)
	if err != nil {
		return err
	}

	_, err = r.query.Delete(&j)
	if err != nil {
		return err
	}

	return nil
}

func (r *Database) Release(queue string, job contractsqueue.Jobs, delay uint) error {
	var j Job
	err := r.query.Where("queue", queue).Where("job", job.Job.Signature()).Where("payloads", job.Args).Order("id desc").FirstOrFail(&j)
	if err != nil {
		return err
	}

	j.AvailableAt = carbon.DateTime{Carbon: carbon.Now().AddSeconds(int(delay))}
	_, err = r.query.Update(&j)
	if err != nil {
		return err
	}

	return nil
}

func (r *Database) Clear(queue string) error {
	_, err := r.query.Where("queue", queue).Delete(&Job{})
	return err
}

func (r *Database) Size(queue string) (uint64, error) {
	var count int64
	err := r.query.Where("queue", queue).Count(&count)
	return uint64(count), err
}
