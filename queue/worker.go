package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/RichardKnop/machinery/v2"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
)

type Worker struct {
	config queue.Config
	job    queue.JobRepository
	json   foundation.Json
	log    log.Log

	connection string
	queue      string
	concurrent int

	currentDelay  time.Duration
	failedJobChan chan FailedJob
	isShutdown    atomic.Bool
	maxDelay      time.Duration
	machinery     *machinery.Worker
	wg            sync.WaitGroup
}

func NewWorker(config queue.Config, job queue.JobRepository, json foundation.Json, log log.Log, connection, queue string, concurrent int) *Worker {
	return &Worker{
		config: config,
		job:    job,
		json:   json,
		log:    log,

		connection: connection,
		queue:      queue,
		concurrent: concurrent,

		currentDelay:  1 * time.Second,
		failedJobChan: make(chan FailedJob, concurrent),
		maxDelay:      32 * time.Second,
	}
}

func (r *Worker) Run() error {
	driver, err := NewDriver(r.connection, r.config)
	if err != nil {
		return err
	}
	if driver.Driver() == queue.DriverSync {
		color.Warningln(errors.QueueDriverSyncNotNeedToRun.Args(r.connection).SetModule(errors.ModuleQueue).Error())
		return nil
	}

	r.isShutdown.Store(false)

	if err := r.RunMachinery(); err != nil {
		return err
	}

	return r.run(driver)
}

// RunMachinery will be removed in v1.17
func (r *Worker) RunMachinery() error {
	instance := NewMachinery(r.config.Config(), r.log, r.job.All(), r.connection, r.queue, r.concurrent)
	if !instance.ExistTasks() {
		return nil
	}

	var (
		worker *machinery.Worker
		err    error
	)

	worker, err = instance.Run()
	if err != nil {
		return err
	}

	r.machinery = worker

	return nil
}

func (r *Worker) Shutdown() error {
	r.isShutdown.Store(true)
	close(r.failedJobChan)

	if r.machinery != nil {
		r.machinery.Quit()
	}

	return nil
}

func (r *Worker) call(task queue.Task) error {
	if !task.Delay.IsZero() {
		time.Sleep(time.Until(task.Delay))
	}

	if err := r.job.Call(task.Job.Signature(), ConvertArgs(task.Args)); err != nil {
		payload, jsonErr := TaskToJson(task, r.json)
		if jsonErr != nil {
			return errors.QueueFailedToConvertTaskToJson.Args(jsonErr, task)
		}

		r.failedJobChan <- FailedJob{
			UUID:       task.UUID,
			Connection: r.connection,
			Queue:      r.queue,
			Payload:    payload,
			Exception:  err.Error(),
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		}
	}

	// TODO: print success log

	return nil
}

func (r *Worker) run(driver queue.Driver) error {
	queueKey := r.config.QueueKey(r.connection, r.queue)

	for i := 0; i < r.concurrent; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				if r.isShutdown.Load() {
					return
				}

				task, err := driver.Pop(queueKey)
				if err != nil {
					if !errors.Is(err, errors.QueueDriverNoJobFound) {
						r.log.Error(errors.QueueDriverFailedToPop.Args(queueKey, err))

						r.currentDelay *= 2
						if r.currentDelay > r.maxDelay {
							r.currentDelay = r.maxDelay
						}
					}

					time.Sleep(r.currentDelay)

					continue
				}

				r.currentDelay = 1 * time.Second

				// the main job should be delayed in the driver
				task.Delay = time.Time{}
				if err := r.call(task); err != nil {
					r.log.Error(err)
					continue
				}

				if len(task.Chain) > 0 {
					for i, chain := range task.Chain {
						chainTask := queue.Task{
							Jobs:  chain,
							UUID:  task.UUID,
							Chain: task.Chain[i+1:],
						}

						if err := r.call(chainTask); err != nil {
							r.log.Error(err)
							continue
						}
					}
				}
			}
		}()
	}

	r.wg.Add(1)

	go func() {
		defer r.wg.Done()
		for job := range r.failedJobChan {
			if _, err := r.config.FailedJobsQuery().Insert(&job); err != nil {
				r.log.Error(errors.QueueFailedToSaveFailedJob.Args(err, job))
			}

			// TODO: print failed log
		}
	}()

	r.wg.Wait()

	return nil
}
