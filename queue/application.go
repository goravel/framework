package queue

import (
	configcontract "github.com/goravel/framework/contracts/config"
	eventcontract "github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config *Config
	event  eventcontract.Instance
	jobs   []queue.Job
}

func NewApplication(config configcontract.Config, event eventcontract.Instance) *Application {
	return &Application{
		config: NewConfig(config),
		event:  event,
	}
}

func (app *Application) Worker(args *queue.Args) queue.Worker {
	defaultConnection := app.config.DefaultConnection()

	if args == nil {
		return NewWorker(app.config, 1, defaultConnection, app.event.GetEvents(), app.jobs, app.config.Queue(defaultConnection, ""))
	}
	if args.Connection == "" {
		args.Connection = defaultConnection
	}

	return NewWorker(app.config, args.Concurrent, args.Connection, app.event.GetEvents(), app.jobs, app.config.Queue(args.Connection, args.Queue))
}

func (app *Application) Register(jobs []queue.Job) {
	app.jobs = append(app.jobs, jobs...)
}

func (app *Application) GetJobs() []queue.Job {
	return app.jobs
}

func (app *Application) Job(job queue.Job, args []queue.Arg) queue.Task {
	return NewTask(app.config, job, args)
}

func (app *Application) Chain(jobs []queue.Jobs) queue.Task {
	return NewChainTask(app.config, jobs)
}
