package queue

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config    queue.Config
	db        db.DB
	jobStorer queue.JobStorer
	json      foundation.Json
	log       log.Log
}

func NewApplication(config queue.Config, db db.DB, job queue.JobStorer, json foundation.Json, log log.Log) *Application {
	return &Application{
		config:    config,
		db:        db,
		jobStorer: job,
		json:      json,
		log:       log,
	}
}

func (r *Application) Connection(name string) (queue.Driver, error) {
	return NewDriverCreator(r.config, r.db, r.jobStorer, r.json, r.log).Create(name)
}

func (r *Application) Chain(jobs []queue.ChainJob) queue.PendingJob {
	return NewPendingChainJob(r.config, r.db, r.jobStorer, r.json, jobs, r.log)
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

func (r *Application) JobStorer() queue.JobStorer {
	return r.jobStorer
}

func (r *Application) Job(job queue.Job, args ...[]queue.Arg) queue.PendingJob {
	return NewPendingJob(r.config, r.db, r.jobStorer, r.json, job, r.log, args...)
}

func (r *Application) Register(jobs []queue.Job) {
	r.jobStorer.Register(jobs)
}

func (r *Application) Worker(payloads ...queue.Args) queue.Worker {
	defaultConnection := r.config.DefaultConnection()
	defaultQueue := r.config.DefaultQueue()
	defaultConcurrent := r.config.DefaultConcurrent()

	if len(payloads) == 0 {
		worker, err := NewWorker(r.config, r.db, r.jobStorer, r.json, r.log, defaultConnection, defaultQueue, defaultConcurrent)
		if err != nil {
			panic(err)
		}
		return worker
	}
	if payloads[0].Connection == "" {
		payloads[0].Connection = defaultConnection
	}
	if payloads[0].Queue == "" {
		payloads[0].Queue = r.config.GetString(fmt.Sprintf("queue.connections.%s.queue", payloads[0].Connection), "default")
	}
	if payloads[0].Concurrent == 0 {
		payloads[0].Concurrent = r.config.GetInt(fmt.Sprintf("queue.connections.%s.concurrent", payloads[0].Connection), 1)
	}

	worker, err := NewWorker(r.config, r.db, r.jobStorer, r.json, r.log, payloads[0].Connection, payloads[0].Queue, payloads[0].Concurrent)
	if err != nil {
		panic(err)
	}

	return worker
}
