package queue

import (
	"time"

	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type DatabaseReservedJob struct {
	db        contractsdb.DB
	job       *models.Job
	jobsTable string
	task      contractsqueue.Task
}

func NewDatabaseReservedJob(job *models.Job, db contractsdb.DB, jobStorer contractsqueue.JobStorer, json contractsfoundation.Json, jobsTable string) (*DatabaseReservedJob, error) {
	task, err := utils.JsonToTask(job.Payload, jobStorer, json)
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

// Attempts returns the current attempt count of the job row.
func (r *DatabaseReservedJob) Attempts() int {
	return r.job.Attempts
}

func (r *DatabaseReservedJob) Delete() error {
	_, err := r.db.Table(r.jobsTable).Where("id", r.job.ID).Delete()

	return err
}

// Release clears the reservation and makes the job available again after the
// delay. Attempts is preserved so the next pop increments it (attempt N+1).
// A single UPDATE keeps the row id stable, unlike Laravel's delete+re-insert.
func (r *DatabaseReservedJob) Release(delay time.Duration) error {
	availableAt := carbon.Now()
	if delay > 0 {
		// Sub-second delays are preserved: time.Time.Add keeps the full
		// Duration (the old int() cast dropped them, releasing the job
		// immediately), and the value is persisted as a full-precision
		// time.Time. SQLite/PostgreSQL store that as-is; MySQL truncates at
		// the column level unless it has fractional precision — the jobs table
		// now uses `DateTimeTz(column, 3)`; existing tables need `datetime(3)`.
		availableAt = carbon.FromStdTime(availableAt.StdTime().Add(delay))
	}

	_, err := r.db.Table(r.jobsTable).Where("id", r.job.ID).Update(map[string]any{
		"reserved_at":  nil,
		"available_at": carbon.NewDateTime(availableAt),
	})

	if err == nil {
		r.job.ReservedAt = nil
	}

	return err
}

func (r *DatabaseReservedJob) Task() contractsqueue.Task {
	return r.task
}
