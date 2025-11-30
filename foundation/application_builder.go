package foundation

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation/configuration"
	"github.com/goravel/framework/support/color"
	"google.golang.org/grpc"
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
	grpcClientInterceptors     map[string][]grpc.UnaryClientInterceptor
	grpcServerInterceptors     []grpc.UnaryServerInterceptor
	jobs                       []queue.Job
	middleware                 func(middleware contractsconfiguration.Middleware)
	migrations                 []schema.Migration
	paths                      func(paths contractsconfiguration.Paths)
	routes                     []func()
	rules                      []validation.Rule
	scheduledEvents            []schedule.Event
	seeders                    []seeder.Seeder
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	// Set custom paths
	if r.paths != nil {
		paths := configuration.NewPaths()
		r.paths(paths)
	}

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

	// Register database migrations
	if len(r.migrations) > 0 {
		schemaFacade := r.app.MakeSchema()
		if schemaFacade == nil {
			color.Errorln("Schema facade not found, please install it first: ./artisan package:install Schema")
		} else {
			schemaFacade.Register(r.migrations)
		}
	}

	// Register database seeders
	if len(r.seeders) > 0 {
		seederFacade := r.app.MakeSeeder()
		if seederFacade == nil {
			color.Errorln("Seeder facade not found, please install it first: ./artisan package:install Seeder")
		} else {
			seederFacade.Register(r.seeders)
		}
	}

	// Register gRPC interceptors
	if len(r.grpcClientInterceptors) > 0 || len(r.grpcServerInterceptors) > 0 {
		grpcFacade := r.app.MakeGrpc()
		if grpcFacade == nil {
			color.Errorln("gRPC facade not found, please install it first: ./artisan package:install Grpc")
		} else {
			if len(r.grpcClientInterceptors) > 0 {
				grpcFacade.UnaryClientInterceptorGroups(r.grpcClientInterceptors)
			}
			if len(r.grpcServerInterceptors) > 0 {
				grpcFacade.UnaryServerInterceptors(r.grpcServerInterceptors)
			}
		}
	}

	// Register jobs
	if len(r.jobs) > 0 {
		queueFacade := r.app.MakeQueue()
		if queueFacade == nil {
			color.Errorln("Queue facade not found, please install it first: ./artisan package:install Queue")
		} else {
			queueFacade.Register(r.jobs)
		}
	}

	// Register validation rules
	if len(r.rules) > 0 {
		validationFacade := r.app.MakeValidation()
		if validationFacade == nil {
			color.Errorln("Validation facade not found, please install it first: ./artisan package:install Validation")
		} else {
			validationFacade.AddRules(r.rules)
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

func (r *ApplicationBuilder) WithGrpcClientInterceptors(groupToInterceptors map[string][]grpc.UnaryClientInterceptor) foundation.ApplicationBuilder {
	r.grpcClientInterceptors = groupToInterceptors

	return r
}

func (r *ApplicationBuilder) WithGrpcServerInterceptors(interceptors []grpc.UnaryServerInterceptor) foundation.ApplicationBuilder {
	r.grpcServerInterceptors = interceptors

	return r
}

func (r *ApplicationBuilder) WithJobs(jobs []queue.Job) foundation.ApplicationBuilder {
	r.jobs = jobs

	return r
}

func (r *ApplicationBuilder) WithMiddleware(fn func(handler contractsconfiguration.Middleware)) foundation.ApplicationBuilder {
	r.middleware = fn

	return r
}

func (r *ApplicationBuilder) WithMigrations(migrations []schema.Migration) foundation.ApplicationBuilder {
	r.migrations = migrations

	return r
}

func (r *ApplicationBuilder) WithPaths(fn func(paths contractsconfiguration.Paths)) foundation.ApplicationBuilder {
	r.paths = fn

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

func (r *ApplicationBuilder) WithRules(rules []validation.Rule) foundation.ApplicationBuilder {
	r.rules = rules

	return r
}

func (r *ApplicationBuilder) WithSchedule(events []schedule.Event) foundation.ApplicationBuilder {
	r.scheduledEvents = events

	return r
}

func (r *ApplicationBuilder) WithSeeders(seeders []seeder.Seeder) foundation.ApplicationBuilder {
	r.seeders = seeders

	return r
}
