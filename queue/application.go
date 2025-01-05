package queue

import (
	configcontract "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config *Config
	job    *JobImpl
}

func NewApplication(config configcontract.Config) *Application {
	return &Application{
		config: NewConfig(config),
		job:    NewJobImpl(),
	}
}

func (app *Application) Worker(payloads ...queue.Args) queue.Worker {
	defaultConnection := app.config.DefaultConnection()

	if len(payloads) == 0 {
		return NewWorker(app.config, 1, defaultConnection, app.config.Queue(defaultConnection, ""), app.job)
	}
	if payloads[0].Connection == "" {
		payloads[0].Connection = defaultConnection
	}
	if payloads[0].Concurrent == 0 {
		payloads[0].Concurrent = 1
	}

	return NewWorker(app.config, payloads[0].Concurrent, payloads[0].Connection, app.config.Queue(payloads[0].Connection, payloads[0].Queue), app.job)
}

func (app *Application) Register(jobs []queue.Job) {
	app.job.Register(jobs)
}

func (app *Application) GetJobs() []queue.Job {
	return app.job.GetJobs()
}

func (app *Application) GetJob(signature string) (queue.Job, error) {
	return app.job.Get(signature)
}

func (app *Application) Job(job queue.Job, args []any) queue.Task {
	return NewTask(app.config, job, args)
}

func (app *Application) Chain(jobs []queue.Jobs) queue.Task {
	return NewChainTask(app.config, jobs)
}
