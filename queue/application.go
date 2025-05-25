package queue

import (
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

func (r *Application) Chain(jobs []queue.ChainJob) queue.PendingJob {
	pendingJob, err := NewPendingChainJob(r.config, r.db, r.jobStorer, r.json, jobs)
	if err != nil {
		panic(err)
	}
	return pendingJob
}

func (r *Application) GetJob(signature string) (queue.Job, error) {
	return r.jobStorer.Get(signature)
}

func (r *Application) GetJobs() []queue.Job {
	return r.jobStorer.All()
}

func (r *Application) Job(job queue.Job, args ...[]queue.Arg) queue.PendingJob {
	pendingJob, err := NewPendingJob(r.config, r.db, r.jobStorer, r.json, job, args...)
	if err != nil {
		panic(err)
	}
	return pendingJob
}

func (r *Application) Register(jobs []queue.Job) {
	r.jobStorer.Register(jobs)
}

func (r *Application) Worker(payloads ...queue.Args) queue.Worker {
	defaultConnection := r.config.DefaultConnection()
	defaultQueue := r.config.DefaultQueue()
	defaultConcurrent := r.config.DefaultConcurrent()

	if len(payloads) == 0 {
		return NewWorker(r.config, r.db, r.jobStorer, r.json, r.log, defaultConnection, defaultQueue, defaultConcurrent)
	}
	if payloads[0].Connection == "" {
		payloads[0].Connection = defaultConnection
	}
	if payloads[0].Queue == "" {
		payloads[0].Queue = defaultQueue
	}
	if payloads[0].Concurrent == 0 {
		payloads[0].Concurrent = defaultConcurrent
	}

	return NewWorker(r.config, r.db, r.jobStorer, r.json, r.log, payloads[0].Connection, payloads[0].Queue, payloads[0].Concurrent)
}
