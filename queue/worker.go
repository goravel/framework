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
	currentDelay  time.Duration
	maxDelay      time.Duration
}

func NewWorker(config queue.Config, concurrent int, connection string, queue string, job queue.JobRepository) *Worker {
	return &Worker{
		concurrent:    concurrent,
		config:        config,
		connection:    connection,
		job:           job,
		queue:         queue,
		failedJobChan: make(chan FailedJob, concurrent),
		currentDelay:  1 * time.Second,
		maxDelay:      32 * time.Second,
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
					if !errors.Is(err, errors.QueueDriverNoJobFound) {
						LogFacade.Error(errors.QueueDriverFailedToPop.Args(r.queue, err))

						r.currentDelay *= 2
						if r.currentDelay > r.maxDelay {
							r.currentDelay = r.maxDelay
						}
					}

					time.Sleep(r.currentDelay)

					continue
				}

				r.currentDelay = 1 * time.Second

				if err = r.job.Call(job.Signature(), args); err != nil {
					r.failedJobChan <- FailedJob{
						UUID:       uuid.New(),
						Connection: r.connection,
						Queue:      r.queue,
						Payload:    args,
						Exception:  err.Error(),
						FailedAt:   carbon.NewLayoutType[carbon.DateTime](carbon.Now()),
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
