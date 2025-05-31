package queue

import (
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type DatabaseReservedJob struct {
	db        contractsdb.DB
	jobRecord *DatabaseJobRecord
	jobsTable string
	task      contractsqueue.Task
}

func NewDatabaseReservedJob(job *DatabaseJobRecord, db contractsdb.DB, jobStorer contractsqueue.JobStorer, json contractsfoundation.Json, jobsTable string) (*DatabaseReservedJob, error) {
	task, err := utils.JsonToTask(job.Payload, jobStorer, json)
	if err != nil {
		return nil, err
	}

	return &DatabaseReservedJob{
		db:        db,
		jobRecord: job,
		jobsTable: jobsTable,
		task:      task,
	}, nil
}

func (r *DatabaseReservedJob) Delete() error {
	_, err := r.db.Table(r.jobsTable).Where("id", r.jobRecord.ID).Delete()

	return err
}

func (r *DatabaseReservedJob) Task() contractsqueue.Task {
	return r.task
}

type DatabaseJobRecord struct {
	ID          uint             `db:"id"`
	Queue       string           `db:"queue"`
	Payload     string           `db:"payload"`
	Attempts    int              `db:"attempts"`
	ReservedAt  *carbon.DateTime `db:"reserved_at"`
	AvailableAt *carbon.DateTime `db:"available_at"`
	CreatedAt   *carbon.DateTime `db:"created_at"`
}

func (r *DatabaseJobRecord) Increment() int {
	r.Attempts++

	return r.Attempts
}

func (r *DatabaseJobRecord) Touch() *carbon.DateTime {
	r.ReservedAt = carbon.NewDateTime(carbon.Now())

	return r.ReservedAt
}
