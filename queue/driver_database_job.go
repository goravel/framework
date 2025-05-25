package queue

import (
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type DatabaseReservedJob struct {
	db        contractsdb.DB
	job       *DatabaseJob
	jobsTable string
	task      contractsqueue.Task
}

func NewDatabaseReservedJob(job *DatabaseJob, db contractsdb.DB, jobStorer contractsqueue.JobStorer, json contractsfoundation.Json, jobsTable string) (*DatabaseReservedJob, error) {
	task, err := JsonToTask(job.Payload, jobStorer, json)
	if err != nil {
		return nil, err
	}

	return &DatabaseReservedJob{
		db:        db,
		job:       job,
		jobsTable: jobsTable,
		task:      task,
	}, nil
}

func (r *DatabaseReservedJob) Delete() error {
	_, err := r.db.Table(r.jobsTable).Where("id", r.job.ID).Delete()

	return err
}

func (r *DatabaseReservedJob) Task() contractsqueue.Task {
	return r.task
}

type DatabaseJob struct {
	ID          uint             `db:"id"`
	Queue       string           `db:"queue"`
	Payload     string           `db:"payload"`
	Attempts    int              `db:"attempts"`
	ReservedAt  *carbon.DateTime `db:"reserved_at"`
	AvailableAt *carbon.DateTime `db:"available_at"`
	CreatedAt   *carbon.DateTime `db:"created_at"`
}

func (r *DatabaseJob) Increment() int {
	r.Attempts++

	return r.Attempts
}

func (r *DatabaseJob) Touch() *carbon.DateTime {
	r.ReservedAt = carbon.NewDateTime(carbon.Now())

	return r.ReservedAt
}
