package event

import (
	"github.com/goravel/framework/contracts/event"
	queuecontract "github.com/goravel/framework/contracts/queue"
)

type Application struct {
	events map[event.Event][]event.Listener
	queue  queuecontract.Queue
}

func NewApplication(queue queuecontract.Queue) *Application {
	return &Application{
		queue: queue,
	}
}

func (app *Application) Register(events map[event.Event][]event.Listener) {
	app.events = events
	var jobs []queuecontract.Job

	for _, listeners := range events {
		for _, listener := range listeners {
			jobs = append(jobs, listener)
		}
	}

	app.queue.Register(jobs)
}

func (app *Application) GetEvents() map[event.Event][]event.Listener {
	return app.events
}

func (app *Application) Job(e event.Event, args []event.Arg) event.Task {
	listeners, ok := app.events[e]
	if !ok {
		listeners = make([]event.Listener, 0)
	}

	return NewTask(app.queue, args, e, listeners)
}
