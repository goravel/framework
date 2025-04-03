package queue

import (
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config queue.Config
	job    queue.JobRepository
	log    log.Log
}

func NewApplication(config queue.Config, job queue.JobRepository, log log.Log) *Application {
	return &Application{
		config: config,
		job:    job,
		log:    log,
	}
}

func (r *Application) Chain(jobs []queue.Jobs) queue.Task {
	return NewChainTask(r.config, jobs)
}

func (r *Application) GetJob(signature string) (queue.Job, error) {
	return r.job.Get(signature)
}

func (r *Application) GetJobs() []queue.Job {
	return r.job.All()
}

func (r *Application) Job(job queue.Job, args ...[]any) queue.Task {
	return NewTask(r.config, job, args...)
}

func (r *Application) Register(jobs []queue.Job) {
	r.job.Register(jobs)
}

func (r *Application) Worker(payloads ...queue.Args) queue.Worker {
	defaultConnection := r.config.DefaultConnection()

	if len(payloads) == 0 {
		return NewWorker(r.config, 1, defaultConnection, r.config.Queue(defaultConnection, ""), r.job, r.log)
	}
	if payloads[0].Connection == "" {
		payloads[0].Connection = defaultConnection
	}
	if payloads[0].Queue == "" {
		payloads[0].Queue = "default"
	}
	if payloads[0].Concurrent == 0 {
		payloads[0].Concurrent = 1
	}

	return NewWorker(r.config, payloads[0].Concurrent, payloads[0].Connection, r.config.Queue(payloads[0].Connection, payloads[0].Queue), r.job, r.log)
}
