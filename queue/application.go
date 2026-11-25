package queue

import (
	"fmt"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config    queue.Config
	cache     cache.Cache
	db        db.DB
	fake      *Fake
	jobStorer queue.JobStorer
	json      foundation.Json
	log       log.Log
}

func NewApplication(config queue.Config, cache cache.Cache, db db.DB, job queue.JobStorer, json foundation.Json, log log.Log) *Application {
	return &Application{
		config:    config,
		cache:     cache,
		db:        db,
		fake:      NewFake(),
		jobStorer: job,
		json:      json,
		log:       log,
	}
}

func (r *Application) Connection(name string) (queue.Driver, error) {
	return NewDriverCreator(r.config, r.cache, r.db, r.jobStorer, r.json, r.log).Create(name)
}

func (r *Application) Chain(jobs []queue.ChainJob) queue.PendingJob {
	pendingJob := NewPendingChainJob(r.config, r.cache, r.db, r.jobStorer, r.json, jobs, r.log)
	if pendingJob != nil {
		pendingJob.fake = r.fake
	}

	return pendingJob
}

func (r *Application) GetJob(signature string) (queue.Job, error) {
	return r.jobStorer.Get(signature)
}

func (r *Application) GetJobs() []queue.Job {
	return r.jobStorer.All()
}

func (r *Application) Failer() queue.Failer {
	return NewFailer(r.config, r.db, r, r.json)
}

// Fake starts faking the jobs, see queue.Queue for the accepted forms.
func (r *Application) Fake(jobs ...any) queue.Fake {
	if r.fake == nil {
		r.fake = NewFake()
	}
	r.fake.enable(jobs)

	return r.fake
}

func (r *Application) JobStorer() queue.JobStorer {
	return r.jobStorer
}

func (r *Application) Job(job queue.Job, args ...[]queue.Arg) queue.PendingJob {
	pendingJob := NewPendingJob(r.config, r.cache, r.db, r.jobStorer, r.json, job, r.log, args...)
	pendingJob.fake = r.fake

	return pendingJob
}

func (r *Application) Register(jobs []queue.Job) {
	r.jobStorer.Register(jobs)
}

func (r *Application) Worker(payloads ...queue.Args) queue.Worker {
	defaultConnection := r.config.DefaultConnection()
	defaultConcurrent := r.config.DefaultConcurrent()

	if len(payloads) == 0 {
		workerQueue := configuredQueue(r.config, defaultConnection)

		worker, err := NewWorker(r.config, r.cache, r.db, r.jobStorer, r.json, r.log, defaultConnection, workerQueue, defaultConcurrent, 1)
		if err != nil {
			panic(err)
		}
		return worker
	}
	if payloads[0].Connection == "" {
		payloads[0].Connection = defaultConnection
	}
	if payloads[0].Queue == "" {
		payloads[0].Queue = configuredQueue(r.config, payloads[0].Connection)
	}
	if payloads[0].Concurrent == 0 {
		payloads[0].Concurrent = r.config.GetInt(fmt.Sprintf("queue.connections.%s.concurrent", payloads[0].Connection), 1)
	}

	worker, err := NewWorker(r.config, r.cache, r.db, r.jobStorer, r.json, r.log, payloads[0].Connection, payloads[0].Queue, payloads[0].Concurrent, payloads[0].Tries)
	if err != nil {
		panic(err)
	}

	return worker
}
