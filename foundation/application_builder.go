package foundation

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/color"
)

func Configure() foundation.ApplicationBuilder {
	return NewApplicationBuilder(App)
}

type ApplicationBuilder struct {
	app              foundation.Application
	config           func()
	eventToListeners map[event.Event][]event.Listener
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	r.app.Boot()

	if r.config != nil {
		r.config()
	}

	if len(r.eventToListeners) > 0 {
		event := r.app.MakeEvent()
		if event == nil {
			color.Errorln("Event facade not found, please install it first: ./artisan package:install Event")
		} else {
			event.Register(r.eventToListeners)
		}
	}

	return r.app
}

func (r *ApplicationBuilder) Run() {
	r.Create().Run()
}

func (r *ApplicationBuilder) WithConfig(fn func()) foundation.ApplicationBuilder {
	r.config = fn

	return r
}

func (r *ApplicationBuilder) WithEvents(eventToListeners map[event.Event][]event.Listener) foundation.ApplicationBuilder {
	r.eventToListeners = eventToListeners

	return r
}
