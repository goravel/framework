package queue

import (
	configcontract "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config *Config
	jobs   []queue.Job
	log    log.Log
}

func NewApplication(config configcontract.Config, log log.Log) *Application {
	return &Application{
		config: NewConfig(config),
		log:    log,
	}
}

func (app *Application) Worker(args ...queue.Args) queue.Worker {
	defaultConnection := app.config.DefaultConnection()

	if len(args) == 0 {
		return NewWorker(app.config, app.log, 1, defaultConnection, app.jobs, app.config.Queue(defaultConnection, ""))
	}

	if args[0].Connection == "" {
		args[0].Connection = defaultConnection
	}

	return NewWorker(app.config, app.log, args[0].Concurrent, args[0].Connection, app.jobs, app.config.Queue(args[0].Connection, args[0].Queue))
}

func (app *Application) Register(jobs []queue.Job) {
	app.jobs = append(app.jobs, jobs...)
}

func (app *Application) GetJobs() []queue.Job {
	return app.jobs
}

func (app *Application) Job(job queue.Job, args []queue.Arg) queue.Task {
	return NewTask(app.config, app.log, job, args)
}

func (app *Application) Chain(jobs []queue.Jobs) queue.Task {
	return NewChainTask(app.config, app.log, jobs)
}
