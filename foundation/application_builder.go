package foundation

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation/configuration"
	"github.com/goravel/framework/support/color"
)

func Setup() foundation.ApplicationBuilder {
	return NewApplicationBuilder(App)
}

type ApplicationBuilder struct {
	app                        foundation.Application
	commands                   []console.Command
	config                     func()
	configuredServiceProviders []foundation.ServiceProvider
	eventToListeners           map[event.Event][]event.Listener
	middleware                 func(middleware contractsconfiguration.Middleware)
	routes                     []func()
	scheduledEvents            []schedule.Event
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

	// Register http middleware
	if r.middleware != nil {
		routeFacade := r.app.MakeRoute()
		if routeFacade == nil {
			color.Errorln("Route facade not found, please install it first: ./artisan package:install Route")
		} else {
			// Set up global middleware
			defaultGlobalMiddleware := routeFacade.GetGlobalMiddleware()
			middleware := configuration.NewMiddleware(defaultGlobalMiddleware)
			r.middleware(middleware)
			routeFacade.SetGlobalMiddleware(middleware.GetGlobalMiddleware())

			// Set up custom recover function
			if recover := middleware.GetRecover(); recover != nil {
				routeFacade.Recover(recover)
			}
		}
	}

	// Register routes
	for _, route := range r.routes {
		route()
	}

	// Register event listeners
	if len(r.eventToListeners) > 0 {
		eventFacade := r.app.MakeEvent()
		if eventFacade == nil {
			color.Errorln("Event facade not found, please install it first: ./artisan package:install Event")
		} else {
			eventFacade.Register(r.eventToListeners)
		}
	}

	// Register commands
	if len(r.commands) > 0 {
		artisanFacade := r.app.MakeArtisan()
		if artisanFacade == nil {
			color.Errorln("Artisan facade not found, please install it first: ./artisan package:install Artisan")
		} else {
			artisanFacade.Register(r.commands)
		}
	}

	// Register scheduled events
	if len(r.scheduledEvents) > 0 {
		scheduleFacade := r.app.MakeSchedule()
		if scheduleFacade == nil {
			color.Errorln("Schedule facade not found, please install it first: ./artisan package:install Schedule")
		} else {
			scheduleFacade.Register(r.scheduledEvents)
		}
	}

	return r.app
}

func (r *ApplicationBuilder) Run() {
	r.Create().Run()
}

func (r *ApplicationBuilder) WithCommands(commands []console.Command) foundation.ApplicationBuilder {
	r.commands = commands

	return r
}

func (r *ApplicationBuilder) WithConfig(config func()) foundation.ApplicationBuilder {
	r.config = config

	return r
}

func (r *ApplicationBuilder) WithEvents(eventToListeners map[event.Event][]event.Listener) foundation.ApplicationBuilder {
	r.eventToListeners = eventToListeners

	return r
}

func (r *ApplicationBuilder) WithMiddleware(fn func(handler contractsconfiguration.Middleware)) foundation.ApplicationBuilder {
	r.middleware = fn

	return r
}

func (r *ApplicationBuilder) WithProviders(providers []foundation.ServiceProvider) foundation.ApplicationBuilder {
	r.configuredServiceProviders = append(r.configuredServiceProviders, providers...)

	return r
}

func (r *ApplicationBuilder) WithRouting(routes []func()) foundation.ApplicationBuilder {
	r.routes = append(r.routes, routes...)

	return r
}

func (r *ApplicationBuilder) WithSchedule(events []schedule.Event) foundation.ApplicationBuilder {
	r.scheduledEvents = events

	return r
}
