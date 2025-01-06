package queue

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type Worker struct {
	concurrent    int
	config        queue.Config
	connection    string
	driver        queue.Driver
	failedJobChan chan FailedJob
	isShutdown    bool
	job           queue.JobRepository
	queue         string
	wg            sync.WaitGroup
}

func NewWorker(config queue.Config, concurrent int, connection string, queue string, job queue.JobRepository) *Worker {
	return &Worker{
		concurrent:    concurrent,
		config:        config,
		connection:    connection,
		job:           job,
		queue:         queue,
		failedJobChan: make(chan FailedJob),
	}
}

func (r *Worker) Run() error {
	r.isShutdown = false

	driver, err := NewDriver(r.connection, r.config)
	if err != nil {
		return err
	}
	if driver.Driver() == queue.DriverSync {
		return errors.QueueDriverSyncNotNeedRun.Args(r.queue)
	}

	for i := 0; i < r.concurrent; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				if r.isShutdown {
					return
				}

				job, args, err := driver.Pop(r.queue)
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				}

				if err = r.job.Call(job.Signature(), args); err != nil {
					r.failedJobChan <- FailedJob{
						UUID:       uuid.New(),
						Connection: r.connection,
						Queue:      r.queue,
						Payload:    args,
						Exception:  err.Error(),
						FailedAt:   carbon.DateTime{Carbon: carbon.Now()},
					}
				}
			}
		}()
	}

	go func() {
		for job := range r.failedJobChan {
			if err = r.config.FailedJobsQuery().Create(&job); err != nil {
				LogFacade.Error(errors.QueueFailedToSaveFailedJob.Args(err))
			}
		}
	}()

	return nil
}

func (r *Worker) Shutdown() error {
	r.isShutdown = true
	r.wg.Wait()
	close(r.failedJobChan)
	return nil
}
