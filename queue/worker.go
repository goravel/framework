package queue

import (
	"sync"
	"sync/atomic"
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
	failedJobChan chan FailedJob
	isShutdown    atomic.Bool
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
		failedJobChan: make(chan FailedJob, concurrent),
	}
}

func (r *Worker) Run() error {
	r.isShutdown.Store(false)

	driver, err := NewDriver(r.connection, r.config)
	if err != nil {
		return err
	}
	if driver.Driver() == queue.DriverSync {
		return errors.QueueDriverSyncNotNeedRun.Args(r.queue)
	}

	// special cases for Machinery
	// TODO: will remove in v1.17
	if driver.Driver() == queue.DriverMachinery {
		return r.runMachinery(driver)
	}

	for i := 0; i < r.concurrent; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				if r.isShutdown.Load() {
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

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for job := range r.failedJobChan {
			if err = r.config.FailedJobsQuery().Create(&job); err != nil {
				LogFacade.Error(errors.QueueFailedToSaveFailedJob.Args(err))
			}
		}
	}()

	r.wg.Wait()

	return nil
}

func (r *Worker) Shutdown() error {
	r.isShutdown.Store(true)
	close(r.failedJobChan)
	return nil
}

// runMachinery is a special case for Machinery
// TODO: will remove in v1.17
func (r *Worker) runMachinery(driver queue.Driver) error {
	m := driver.(*Machinery)
	return m.Run(r.job.All(), r.queue, r.concurrent)
}
