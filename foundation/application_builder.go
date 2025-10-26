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
	app                        foundation.Application
	config                     func()
	eventToListeners           map[event.Event][]event.Listener
	configuredServiceProviders []foundation.ServiceProvider
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	if len(r.configuredServiceProviders) > 0 {
		r.app.AddServiceProviders(r.configuredServiceProviders)
	}

	r.app.Boot()

	if r.config != nil {
		r.config()
	}

	if len(r.eventToListeners) > 0 {
		evt := r.app.MakeEvent()
		if evt == nil {
			color.Errorln("Event facade not found, please install it first: ./artisan package:install Event")
		} else {
			evt.Register(r.eventToListeners)
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

func (r *ApplicationBuilder) WithProviders(providers []foundation.ServiceProvider) foundation.ApplicationBuilder {
	r.configuredServiceProviders = append(r.configuredServiceProviders, providers...)

	return r
}

func (r *ApplicationBuilder) WithEvents(eventToListeners map[event.Event][]event.Listener) foundation.ApplicationBuilder {
	r.eventToListeners = eventToListeners

	return r
}
