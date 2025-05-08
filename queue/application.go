package queue

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config queue.Config
	job    queue.JobRepository
	json   foundation.Json
	log    log.Log
}

func NewApplication(config queue.Config, job queue.JobRepository, json foundation.Json, log log.Log) *Application {
	return &Application{
		config: config,
		job:    job,
		json:   json,
		log:    log,
	}
}

func (r *Application) Chain(jobs []queue.Jobs) queue.PendingJob {
	return NewPendingChainJob(r.config, jobs)
}

func (r *Application) GetJob(signature string) (queue.Job, error) {
	return r.job.Get(signature)
}

func (r *Application) GetJobs() []queue.Job {
	return r.job.All()
}

func (r *Application) Job(job queue.Job, args ...[]queue.Arg) queue.PendingJob {
	return NewPendingJob(r.config, job, args...)
}

func (r *Application) Register(jobs []queue.Job) {
	r.job.Register(jobs)
}

func (r *Application) Worker(payloads ...queue.Args) queue.Worker {
	defaultConnection := r.config.DefaultConnection()
	defaultQueue := r.config.DefaultQueue()
	defaultConcurrent := r.config.DefaultConcurrent()

	if len(payloads) == 0 {
		return NewWorker(r.config, r.job, r.json, r.log, defaultConnection, defaultQueue, defaultConcurrent)
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

	return NewWorker(r.config, r.job, r.json, r.log, payloads[0].Connection, payloads[0].Queue, payloads[0].Concurrent)
}
