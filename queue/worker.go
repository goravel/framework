package queue

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type Worker struct {
	concurrent    int
	driver        *DriverImpl
	job           *JobImpl
	failedJobs    orm.Query
	queue         string
	failedJobChan chan FailedJob
	isShutdown    bool
}

func NewWorker(config *Config, concurrent int, connection string, queue string, job *JobImpl) *Worker {
	return &Worker{
		concurrent:    concurrent,
		driver:        NewDriverImpl(connection, config),
		job:           job,
		failedJobs:    config.FailedJobsQuery(),
		queue:         queue,
		failedJobChan: make(chan FailedJob),
	}
}

func (r *Worker) Run() error {
	r.isShutdown = false

	driver, err := r.driver.New()
	if err != nil {
		return err
	}
	if driver.Driver() == DriverSync {
		return fmt.Errorf("queue %s driver not need run", r.queue)
	}

	for i := 0; i < r.concurrent; i++ {
		go func() {
			for {
				if r.isShutdown {
					return
				}

				job, args, err := driver.Pop(r.queue)
				if err != nil {
					// This error not need to be reported.
					// It is usually caused by the queue being empty.
					time.Sleep(1 * time.Second)
					continue
				}

				if err = r.job.Call(job.Signature(), args); err != nil {
					r.failedJobChan <- FailedJob{
						Queue:     r.queue,
						Signature: job.Signature(),
						Payloads:  args,
						Exception: err.Error(),
						FailedAt:  carbon.DateTime{Carbon: carbon.Now()},
					}
				}
			}
		}()
	}

	go func() {
		for job := range r.failedJobChan {
			_ = r.failedJobs.Create(&job)
		}
	}()

	return nil
}

func (r *Worker) Shutdown() error {
	r.isShutdown = true
	return nil
}
