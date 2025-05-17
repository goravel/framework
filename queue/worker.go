package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/RichardKnop/machinery/v2"

	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/console"
)

type Worker struct {
	config queue.Config
	db     db.DB
	job    queue.JobRepository
	json   foundation.Json
	log    log.Log

	connection string
	queue      string
	concurrent int
	debug      bool

	currentDelay  time.Duration
	failedJobChan chan FailedJob
	isShutdown    atomic.Bool
	maxDelay      time.Duration
	machinery     *machinery.Worker
	wg            sync.WaitGroup
}

func NewWorker(config queue.Config, db db.DB, job queue.JobRepository, json foundation.Json, log log.Log, connection, queue string, concurrent int) *Worker {
	return &Worker{
		config: config,
		db:     db,
		job:    job,
		json:   json,
		log:    log,

		connection: connection,
		queue:      queue,
		concurrent: concurrent,
		debug:      config.Debug(),

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
	instance := NewMachinery(r.config, r.log, r.job.All(), r.connection, r.queue, r.concurrent)
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
	r.printRunningLog(task)

	if !task.Delay.IsZero() {
		time.Sleep(time.Until(task.Delay))
	}

	now := carbon.Now()
	err := r.job.Call(task.Job.Signature(), ConvertArgs(task.Args))
	duration := carbon.Now().DiffAbsInDuration(now).String()

	if err != nil {
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

		r.printFailedLog(task, duration)

		return nil
	}

	r.printSuccessLog(task, duration)

	return nil
}

func (r *Worker) logFailedJob(job FailedJob) {
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

	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeString())
	status := "<fg=yellow;op=bold>RUNNING</>"
	first := datetime + " " + task.Job.Signature()
	second := status

	color.Default().Println(console.TwoColumnDetail(first, second))
}

func (r *Worker) printSuccessLog(task queue.Task, duration string) {
	if !r.debug {
		return
	}

	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeString())
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

	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeString())
	status := "<fg=red;op=bold>FAIL</>"
	duration = color.Gray().Sprint(duration)
	first := datetime + " " + task.Job.Signature()
	second := duration + " " + status

	color.Default().Println(console.TwoColumnDetail(first, second))
}

func (r *Worker) run(driver queue.Driver) error {
	if r.debug {
		color.Infoln(errors.QueueProcessingJobs.Args(r.connection, r.queue))
	}

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
			r.logFailedJob(job)
		}
	}()

	r.wg.Wait()

	return nil
}
