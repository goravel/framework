package events

import (
	"github.com/goravel/framework/contracts/events"
	"github.com/goravel/framework/events/support"
)

type Application struct {
	events map[events.Event][]events.Listener
}

func (app *Application) Register(events map[events.Event][]events.Listener) {
	app.events = events
}

func (app *Application) GetEvents() map[events.Event][]events.Listener {
	return app.events
}

func (app *Application) Job(event events.Event, args []events.Arg) events.Task {
	return &support.Task{
		Events: app.events,
		Event:  event,
		Args:   args,
	}
}
