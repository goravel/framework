package queue

import (
	configcontract "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/queue"
)

// JobRegistry is a map to store all registered jobs.
var JobRegistry = make(map[string]queue.Job)

type Application struct {
	config *Config
}

func NewApplication(config configcontract.Config) *Application {
	return &Application{
		config: NewConfig(config),
	}
}

func (app *Application) Worker(args *queue.Args) queue.Worker {
	defaultConnection := app.config.DefaultConnection()

	if args == nil {
		return NewWorker(app.config, 1, defaultConnection, app.config.Queue(defaultConnection, ""))
	}
	if args.Connection == "" {
		args.Connection = defaultConnection
	}
	if args.Concurrent == 0 {
		args.Concurrent = 1
	}

	return NewWorker(app.config, args.Concurrent, args.Connection, app.config.Queue(args.Connection, args.Queue))
}

func (app *Application) Register(jobs []queue.Job) error {
	if err := Register(jobs); err != nil {
		return err
	}

	return nil
}

func (app *Application) GetJobs() []queue.Job {
	var jobs []queue.Job
	for _, job := range JobRegistry {
		jobs = append(jobs, job)
	}

	return jobs
}

func (app *Application) Job(job queue.Job, args []any) queue.Task {
	return NewTask(app.config, job, args)
}

func (app *Application) Chain(jobs []queue.Jobs) queue.Task {
	return NewChainTask(app.config, jobs)
}
