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
	configuredServiceProviders []foundation.ServiceProvider
	eventToListeners           map[event.Event][]event.Listener
	routes                     []func()
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	// Register and boot custom service providers
	r.app.AddServiceProviders(r.configuredServiceProviders)
	r.app.Boot()

	// Apply custom configuration
	if r.config != nil {
		r.config()
	}

	// Register routes
	for _, route := range r.routes {
		route()
	}

	// Register event listeners
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

func (r *ApplicationBuilder) WithConfig(config func()) foundation.ApplicationBuilder {
	r.config = config

	return r
}

func (r *ApplicationBuilder) WithEvents(eventToListeners map[event.Event][]event.Listener) foundation.ApplicationBuilder {
	r.eventToListeners = eventToListeners

	return r
}

func (r *ApplicationBuilder) WithProviders(providers []foundation.ServiceProvider) foundation.ApplicationBuilder {
	r.configuredServiceProviders = append(r.configuredServiceProviders, providers...)

	return r
}

func (r *ApplicationBuilder) WithRouting(routes ...func()) foundation.ApplicationBuilder {
	r.routes = append(r.routes, routes...)

	return r
}
