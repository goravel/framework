package queue

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue/support"
)

type Application struct {
	jobs []queue.Job
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Worker(args *queue.Args) queue.Worker {
	if args == nil {
		connection := facades.Config.GetString("queue.default")

		return &support.Worker{
			Connection: connection,
			Queue:      support.GetQueueName(connection, ""),
			Concurrent: 1,
		}
	}

	return &support.Worker{
		Connection: args.Connection,
		Queue:      support.GetQueueName(args.Connection, args.Queue),
		Concurrent: args.Concurrent,
	}
}

func (app *Application) Register(jobs []queue.Job) {
	app.jobs = append(app.jobs, jobs...)
}

func (app *Application) GetJobs() []queue.Job {
	return app.jobs
}

func (app *Application) Job(job queue.Job, args []queue.Arg) queue.Task {
	return &support.Task{
		Job:  job,
		Args: args,
	}
}

func (app *Application) Chain(jobs []queue.Jobs) queue.Task {
	return &support.Task{
		Jobs:  jobs,
		Chain: true,
	}
}
