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
}

func (app *Application) GetEvents() map[event.Event][]event.Listener {
	return app.events
}

func (app *Application) Job(event event.Event, args []event.Arg) event.Task {
	listeners, _ := app.events[event]

	return NewTask(app.queue, args, event, listeners)
}
