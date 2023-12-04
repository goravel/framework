package queue

import (
	"github.com/goravel/framework/contracts/database/orm"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type Database struct {
	connection string
	jobs       orm.Query
}

func NewDatabase(connection string, jobsOrm orm.Query) *Database {
	return &Database{
		connection: connection,
		jobs:       jobsOrm,
	}
}

func (r *Database) ConnectionName() string {
	return r.connection
}

func (r *Database) DriverName() string {
	return DriverDatabase
}

func (r *Database) Push(job contractsqueue.Job, payloads []any, queue string) error {
	var j Job
	j.Queue = queue
	j.Job = job.Signature()
	j.Payloads = payloads
	j.AvailableAt = carbon.DateTime{Carbon: carbon.Now()}
	j.CreatedAt = carbon.DateTime{Carbon: carbon.Now()}

	return r.jobs.Create(&j)
}

func (r *Database) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	var j []Job
	for _, job := range jobs {
		var jj Job
		jj.Queue = queue
		jj.Job = job.Job.Signature()
		jj.Payloads = job.Payloads
		jj.AvailableAt = carbon.DateTime{Carbon: carbon.Now().AddSeconds(int(job.Delay))}
		jj.CreatedAt = carbon.DateTime{Carbon: carbon.Now()}
		j = append(j, jj)
	}

	return r.jobs.Create(&j)
}

func (r *Database) Later(delay uint, job contractsqueue.Job, payloads []any, queue string) error {
	var j Job
	j.Queue = queue
	j.Job = job.Signature()
	j.Payloads = payloads
	j.AvailableAt = carbon.DateTime{Carbon: carbon.Now().AddSeconds(int(delay))}
	j.CreatedAt = carbon.DateTime{Carbon: carbon.Now()}

	return r.jobs.Create(&j)
}

func (r *Database) Pop(q string) (contractsqueue.Job, []any, error) {
	var job Job
	err := r.jobs.Where("queue", q).Where("reserved_at", nil).Where("available_at", "<=", carbon.DateTime{Carbon: carbon.Now()}).Order("id asc").First(&job)

	return job, job.Payloads, err
}

func (r *Database) Delete(queue string, job contractsqueue.Jobs) error {
	var j Job
	err := r.jobs.Where("queue", queue).Where("job", job.Job.Signature()).Where("payloads", job.Payloads).Order("id desc").FirstOrFail(&j)
	if err != nil {
		return err
	}

	_, err = r.jobs.Delete(&j)
	if err != nil {
		return err
	}

	return nil
}

func (r *Database) Release(queue string, job contractsqueue.Jobs, delay uint) error {
	var j Job
	err := r.jobs.Where("queue", queue).Where("job", job.Job.Signature()).Where("payloads", job.Payloads).Order("id desc").FirstOrFail(&j)
	if err != nil {
		return err
	}

	j.AvailableAt = carbon.DateTime{Carbon: carbon.Now().AddSeconds(int(delay))}
	_, err = r.jobs.Update(&j)
	if err != nil {
		return err
	}

	return nil
}

func (r *Database) Clear(queue string) error {
	_, err := r.jobs.Where("queue", queue).Delete(&Job{})
	return err
}

func (r *Database) Size(queue string) (uint64, error) {
	var count int64
	err := r.jobs.Where("queue", queue).Count(&count)
	return uint64(count), err
}
