package queue

import (
	"time"

	"github.com/google/uuid"
	contractscache "github.com/goravel/framework/contracts/cache"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractslog "github.com/goravel/framework/contracts/log"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/carbon"
)

type PendingJob struct {
	cache         contractscache.Cache
	connection    string
	driverCreator contractsqueue.DriverCreator
	delay         time.Time
	queue         string
	task          contractsqueue.Task
}

func NewPendingJob(
	config contractsqueue.Config,
	cache contractscache.Cache,
	db contractsdb.DB,
	jobStorer contractsqueue.JobStorer,
	json contractsfoundation.Json,
	job contractsqueue.Job,
	log contractslog.Log,
	args ...[]contractsqueue.Arg) *PendingJob {
	var arg []contractsqueue.Arg
	if len(args) > 0 {
		arg = args[0]
	}

	connection := config.DefaultConnection()
	queue := config.DefaultQueue()

	return &PendingJob{
		cache:         cache,
		connection:    connection,
		driverCreator: NewDriverCreator(config, cache, db, jobStorer, json, log),
		queue:         queue,
		task: contractsqueue.Task{
			UUID: uuid.New().String(),
			ChainJob: contractsqueue.ChainJob{
				Job:  job,
				Args: arg,
			},
		},
	}
}

func NewPendingChainJob(
	config contractsqueue.Config,
	cache contractscache.Cache,
	db contractsdb.DB,
	jobStorer contractsqueue.JobStorer,
	json contractsfoundation.Json,
	jobs []contractsqueue.ChainJob,
	log contractslog.Log,
) *PendingJob {
	if len(jobs) == 0 {
		return nil
	}

	var chain []contractsqueue.ChainJob
	for _, job := range jobs[1:] {
		chain = append(chain, contractsqueue.ChainJob{
			Job:   job.Job,
			Args:  job.Args,
			Delay: job.Delay,
		})
	}

	job := contractsqueue.ChainJob{
		Job:   jobs[0].Job,
		Args:  jobs[0].Args,
		Delay: jobs[0].Delay,
	}

	connection := config.DefaultConnection()
	queue := config.DefaultQueue()

	return &PendingJob{
		cache:         cache,
		connection:    connection,
		driverCreator: NewDriverCreator(config, cache, db, jobStorer, json, log),
		queue:         queue,
		task: contractsqueue.Task{
			UUID:     uuid.New().String(),
			ChainJob: job,
			Chain:    chain,
		},
	}
}

// Delay sets a delay time for the task
func (r *PendingJob) Delay(delay time.Time) contractsqueue.PendingJob {
	r.delay = delay
	return r
}

// Dispatch dispatches the task
func (r *PendingJob) Dispatch() error {
	driver, err := r.driverCreator.Create(r.connection)
	if err != nil {
		return err
	}

	r.recalculateDelay()

	acquired, err := acquireUniqueLock(r.cache, r.task)
	if err != nil || !acquired {
		return err
	}

	err = driver.Push(r.task, r.queue)
	// The sync driver has run the job already, and a job that was not pushed
	// will not run.
	if isUniqueJob(r.task) && (err != nil || driver.Driver() == contractsqueue.DriverSync) {
		releaseUniqueLock(r.cache, r.task)
	}

	return err
}

// DispatchSync dispatches the task synchronously
func (r *PendingJob) DispatchSync() error {
	syncDriver := NewSync()

	r.recalculateDelay()

	return syncDriver.Push(r.task, r.queue)
}

// OnConnection sets the connection name
func (r *PendingJob) OnConnection(connection string) contractsqueue.PendingJob {
	r.connection = connection
	return r
}

// OnQueue sets the single queue to which the task is dispatched.
func (r *PendingJob) OnQueue(queue string) contractsqueue.PendingJob {
	r.queue = queue
	return r
}

func (r *PendingJob) recalculateDelay() {
	if !r.delay.IsZero() {
		if !r.task.Delay.IsZero() {
			r.task.Delay = r.task.Delay.Add(carbon.Now().DiffAbsInDuration(carbon.FromStdTime(r.delay)))
		} else {
			r.task.Delay = r.delay
		}
	}
}
