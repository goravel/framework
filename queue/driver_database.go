package queue

import (
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type Database struct {
	connection string
	db         orm.Orm
}

func NewDatabase(connection string, db orm.Orm) *Database {
	return &Database{
		connection: connection,
		db:         db,
	}
}

func (receiver Database) ConnectionName() string {
	return receiver.connection
}

func (receiver Database) Push(job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	var j Job
	j.Queue = queue
	j.Job = job.Signature()
	j.Arg = args
	j.AvailableAt = &carbon.DateTime{Carbon: carbon.Now()}
	j.CreatedAt = &carbon.DateTime{Carbon: carbon.Now()}

	return receiver.db.Query().Create(&j)
}

func (receiver Database) Bulk(jobs []contractsqueue.Jobs, queue string) error {
	var j []Job
	for _, job := range jobs {
		var jj Job
		jj.Queue = queue
		jj.Job = job.Job.Signature()
		jj.Arg = job.Args
		jj.AvailableAt = &carbon.DateTime{Carbon: carbon.Now()}
		jj.CreatedAt = &carbon.DateTime{Carbon: carbon.Now()}
		j = append(j, jj)
	}

	return receiver.db.Query().Create(&j)
}

func (receiver Database) Later(delay int, job contractsqueue.Job, args []contractsqueue.Arg, queue string) error {
	var j Job
	j.Queue = queue
	j.Job = job.Signature()
	j.Arg = args
	j.AvailableAt = &carbon.DateTime{Carbon: carbon.Now().AddSeconds(delay)}
	j.CreatedAt = &carbon.DateTime{Carbon: carbon.Now()}

	return receiver.db.Query().Create(&j)
}

func (receiver Database) Pop(q string) (contractsqueue.Job, []contractsqueue.Arg, error) {
	var job Job
	err := receiver.db.Query().Model(Job{}).Where("queue", q).Where("reserved_at", nil).First(&job)

	return job, job.Arg, err
}

func (receiver Database) Delete(queue string, job contractsqueue.Job) error {
	var j Job
	err := receiver.db.Query().Model(Job{}).Where("queue", queue).Where("job", job.Signature()).First(&j)
	if err != nil {
		_, err = receiver.db.Query().Delete(&j)
		if err != nil {
			return err
		}
	}

	return nil
}

func (receiver Database) Release(queue string, job contractsqueue.Job, delay int) error {
	var j Job
	err := receiver.db.Query().Model(Job{}).Where("queue", queue).Where("job", job.Signature()).First(&j)
	if err != nil {
		j.ReservedAt = &carbon.DateTime{Carbon: carbon.Now().AddSeconds(delay)}
		_, err = receiver.db.Query().Update(&j)
		if err != nil {
			return err
		}
	}

	return nil
}

func (receiver Database) Clear(queue string) error {
	var j []Job
	err := receiver.db.Query().Model(Job{}).Where("queue", queue).Find(&j)
	if err != nil {
		_, err = receiver.db.Query().Delete(&j)
		if err != nil {
			return err
		}
	}

	return nil
}

func (receiver Database) Size(queue string) (int64, error) {
	var count int64
	err := receiver.db.Query().Model(Job{}).Where("queue", queue).Count(&count)
	return count, err
}

func (receiver Database) Server(concurrent int, q string) {
	for i := 0; i < concurrent; i++ {
		go func() {
			for {
				job, args, err := receiver.Pop(q)
				if err != nil {
					continue
				}

				err = Call(job.Signature(), args)
				if err != nil {
					continue
				}

				time.Sleep(time.Second)
			}
		}()
	}
}
