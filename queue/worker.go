package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/console"
)

// TODO make these constants configurable in the future if needed
const (
	receiveTimeout = 5 * time.Second
	receiveBackoff = 100 * time.Millisecond
)

type Worker struct {
	config queue.Config
	db     db.DB
	driver queue.Driver
	job    queue.JobStorer
	json   foundation.Json
	log    log.Log

	failedJobChan chan models.FailedJob

	connection     string
	queue          string
	jobWg          sync.WaitGroup
	failedJobWg    sync.WaitGroup
	concurrent     int
	tries          int
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc

	currentDelay time.Duration
	maxDelay     time.Duration
	isShutdown   atomic.Bool
	debug        bool
}

func NewWorker(config queue.Config, cache cache.Cache, db db.DB, job queue.JobStorer, json foundation.Json, log log.Log, connection, queue string, concurrent, tries int) (*Worker, error) {
	driverCreator := NewDriverCreator(config, cache, db, job, json, log)
	driver, err := driverCreator.Create(connection)
	if err != nil {
		return nil, err
	}

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	return &Worker{
		config: config,
		db:     db,
		driver: driver,
		job:    job,
		json:   json,
		log:    log,

		connection:     connection,
		queue:          queue,
		concurrent:     concurrent,
		tries:          tries,
		debug:          config.Debug(),
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,

		currentDelay:  1 * time.Second,
		failedJobChan: make(chan models.FailedJob, concurrent),
		maxDelay:      32 * time.Second,
	}, nil
}

func (r *Worker) Run() error {
	if r.driver.Driver() == queue.DriverSync {
		color.Warningln(errors.QueueDriverSyncNotNeedToRun.Args(r.connection).SetModule(errors.ModuleQueue).Error())
		return nil
	}

	r.isShutdown.Store(false)

	return r.run()
}

func (r *Worker) Shutdown() error {
	r.isShutdown.Store(true)
	r.shutdownCancel()

	// Wait for all worker goroutines to finish processing current tasks
	r.jobWg.Wait()

	// Close the failed job channel to allow the failed job processor goroutine to exit
	close(r.failedJobChan)

	// Wait for the failed job processor goroutine to finish
	r.failedJobWg.Wait()

	return nil
}

func (r *Worker) call(task queue.Task) error {
	tries := 1
	r.printRunningLog(task)

	for {
		if !task.Delay.IsZero() {
			time.Sleep(carbon.FromStdTime(task.Delay).DiffAbsInDuration())
		}

		now := carbon.Now()
		err := r.job.Call(task.Job.Signature(), utils.ConvertArgs(task.Args))
		duration := now.DiffAbsInDuration().String()

		if err == nil {
			r.printSuccessLog(task, duration)
			return nil
		}

		shouldRetry := false
		var delay time.Duration = 0

		if jobWithShouldRetry, ok := task.Job.(queue.JobWithShouldRetry); ok {
			shouldRetry, delay = jobWithShouldRetry.ShouldRetry(err, tries)
		} else {
			shouldRetry = tries < r.tries /* || r.tries == 0 */ // Currently, we do not support unlimited retries, see https://github.com/goravel/framework/pull/1123#discussion_r2194272829
		}

		if shouldRetry {
			if delay > 0 {
				time.Sleep(delay)
			}
			tries++
			continue
		}

		payload, jsonErr := utils.TaskToJson(task, r.json)
		if jsonErr != nil {
			return errors.QueueFailedToConvertTaskToJson.Args(jsonErr, task)
		}

		r.failedJobChan <- models.FailedJob{
			UUID:       task.UUID,
			Connection: r.connection,
			Queue:      r.queue,
			Payload:    payload,
			Exception:  err.Error(),
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		}

		r.printFailedLog(task, duration)

		return errors.QueueFailedToCallJob
	}
}

func (r *Worker) logFailedJob(job models.FailedJob) {
	failedDatabase := r.config.FailedDatabase()
	failedTable := r.config.FailedTable()

	isDbDisabled := r.db == nil || failedDatabase == "" || failedTable == ""
	if isDbDisabled {
		r.log.Error(errors.QueueJobFailed.Args(job))
		return
	}

	_, err := r.db.Connection(failedDatabase).Table(failedTable).Insert(&job)
	if err != nil {
		r.log.Error(errors.QueueFailedToSaveFailedJob.Args(err, job))
	}
}

func (r *Worker) printRunningLog(task queue.Task) {
	if !r.debug {
		return
	}

	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeMilliString())
	status := "<fg=yellow;op=bold>RUNNING</>"
	first := datetime + " " + task.Job.Signature()
	second := status

	color.Default().Println(console.TwoColumnDetail(first, second))
}

func (r *Worker) printSuccessLog(task queue.Task, duration string) {
	if !r.debug {
		return
	}

	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeMilliString())
	status := "<fg=green;op=bold>DONE</>"
	duration = color.Gray().Sprint(duration)
	first := datetime + " " + task.Job.Signature()
	second := duration + " " + status

	color.Default().Println(console.TwoColumnDetail(first, second))
}

func (r *Worker) printFailedLog(task queue.Task, duration string) {
	if !r.debug {
		return
	}

	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeMilliString())
	status := "<fg=red;op=bold>FAIL</>"
	duration = color.Gray().Sprint(duration)
	first := datetime + " " + task.Job.Signature()
	second := duration + " " + status

	color.Default().Println(console.TwoColumnDetail(first, second))
}

func (r *Worker) run() error {
	if r.debug {
		color.Infoln(errors.QueueProcessingJobs.Args(r.connection, r.queue).Error())
	}

	if receiver, ok := r.driver.(queue.DriverWithReceive); ok {
		return r.runWithReceive(receiver)
	}

	return r.runWithPop()
}

func (r *Worker) runWithPop() error {
	r.failedJobWg.Add(1)
	go func() {
		defer r.failedJobWg.Done()
		for job := range r.failedJobChan {
			r.logFailedJob(job)
		}
	}()

	for i := 0; i < r.concurrent; i++ {
		r.jobWg.Add(1)
		go func() {
			defer r.jobWg.Done()
			for {
				if r.isShutdown.Load() {
					return
				}

				reservedJob, err := r.driver.Pop(r.queue)
				if err != nil {
					if !errors.Is(err, errors.QueueDriverNoJobFound) {
						r.log.Error(errors.QueueDriverFailedToPop.Args(r.queue, err))

						r.currentDelay *= 2
						if r.currentDelay > r.maxDelay {
							r.currentDelay = r.maxDelay
						}
					}

					time.Sleep(r.currentDelay)

					continue
				}

				r.currentDelay = 1 * time.Second
				r.processReservedJob(reservedJob)
			}
		}()
	}

	r.jobWg.Wait()

	return nil
}

func (r *Worker) runWithReceive(receiver queue.DriverWithReceive) error {
	r.failedJobWg.Add(1)
	go func() {
		defer r.failedJobWg.Done()
		for job := range r.failedJobChan {
			r.logFailedJob(job)
		}
	}()

	r.jobWg.Add(1)
	defer r.jobWg.Done()

	currentDelay := receiveBackoff

	for {
		if r.isShutdown.Load() {
			return nil
		}

		ctx, cancel := context.WithTimeout(r.shutdownCtx, receiveTimeout)
		jobs, err := receiver.Receive(ctx, r.queue, r.concurrent)
		cancel()

		if err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				r.log.Error(errors.QueueDriverFailedToReceive.Args(r.queue, err))

				currentDelay *= 2
				if currentDelay > r.maxDelay {
					currentDelay = r.maxDelay
				}
			}

			time.Sleep(currentDelay)
			continue
		}

		if len(jobs) == 0 {
			time.Sleep(currentDelay)
			continue
		}

		currentDelay = receiveBackoff

		var wg sync.WaitGroup
		for _, reservedJob := range jobs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r.processReservedJob(reservedJob)
			}()
		}
		wg.Wait()
	}
}

func (r *Worker) processReservedJob(reservedJob queue.ReservedJob) {
	task := reservedJob.Task()

	if err := r.call(task); err != nil {
		if !errors.Is(err, errors.QueueFailedToCallJob) {
			r.log.Error(err)
		}

		if err = reservedJob.Delete(); err != nil {
			r.log.Error(errors.QueueFailedToDeleteReservedJob.Args(reservedJob, err))
		}

		return
	}

	if len(task.Chain) > 0 {
		for i, chain := range task.Chain {
			chainTask := queue.Task{
				ChainJob: chain,
				UUID:     task.UUID,
				Chain:    task.Chain[i+1:],
			}

			if err := r.call(chainTask); err != nil {
				if !errors.Is(err, errors.QueueFailedToCallJob) {
					r.log.Error(err)
				}
				break
			}
		}
	}

	if err := reservedJob.Delete(); err != nil {
		r.log.Error(errors.QueueFailedToDeleteReservedJob.Args(reservedJob, err))
	}
}
