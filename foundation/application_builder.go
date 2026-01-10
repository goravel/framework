package foundation

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

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
)

func Setup() foundation.ApplicationBuilder {
	return NewApplicationBuilder(App)
}

type ApplicationBuilder struct {
	app                        foundation.Application
	callback                   func()
	commands                   func() []console.Command
	config                     func()
	configuredServiceProviders func() []foundation.ServiceProvider
	eventToListeners           func() map[event.Event][]event.Listener
	filters                    func() []validation.Filter
	grpcClientInterceptors     func() map[string][]grpc.UnaryClientInterceptor
	grpcClientStatsHandlers    func() map[string][]stats.Handler
	grpcServerInterceptors     func() []grpc.UnaryServerInterceptor
	grpcServerStatsHandlers    func() []stats.Handler
	jobs                       func() []queue.Job
	middleware                 func(middleware contractsconfiguration.Middleware)
	migrations                 func() []schema.Migration
	paths                      func(paths contractsconfiguration.Paths)
	routes                     func()
	rules                      func() []validation.Rule
	schedule                   func() []schedule.Event
	seeders                    func() []seeder.Seeder
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	r.configurePaths()
	r.configureServiceProviders()
	r.registerServiceProviders()
	r.configureCustomConfig()
	r.configureMiddleware()
	r.configureEventListeners()
	r.configureCommands()
	r.configureSchedule()
	r.configureMigrations()
	r.configureSeeders()
	r.configureGrpc()
	r.configureJobs()
	r.configureValidation()
	r.configureRoutes()
	r.configureCallback()
	r.bootServiceProviders()

	return r.app
}

func (r *ApplicationBuilder) Start() foundation.Application {
	return r.Create().Start()
}

func (r *ApplicationBuilder) WithCallback(callback func()) foundation.ApplicationBuilder {
	r.callback = callback

	return r
}

func (r *ApplicationBuilder) WithCommands(fn func() []console.Command) foundation.ApplicationBuilder {
	r.commands = fn

	return r
}

func (r *ApplicationBuilder) WithConfig(fn func()) foundation.ApplicationBuilder {
	r.config = fn

	return r
}

func (r *ApplicationBuilder) WithEvents(fn func() map[event.Event][]event.Listener) foundation.ApplicationBuilder {
	r.eventToListeners = fn

	return r
}

func (r *ApplicationBuilder) WithFilters(fn func() []validation.Filter) foundation.ApplicationBuilder {
	r.filters = fn

	return r
}

func (r *ApplicationBuilder) WithGrpcClientInterceptors(fn func() map[string][]grpc.UnaryClientInterceptor) foundation.ApplicationBuilder {
	r.grpcClientInterceptors = fn

	return r
}

func (r *ApplicationBuilder) WithGrpcClientStatsHandlers(fn func() map[string][]stats.Handler) foundation.ApplicationBuilder {
	r.grpcClientStatsHandlers = fn

	return r
}

func (r *ApplicationBuilder) WithGrpcServerInterceptors(fn func() []grpc.UnaryServerInterceptor) foundation.ApplicationBuilder {
	r.grpcServerInterceptors = fn

	return r
}

func (r *ApplicationBuilder) WithGrpcServerStatsHandlers(fn func() []stats.Handler) foundation.ApplicationBuilder {
	r.grpcServerStatsHandlers = fn

	return r
}

func (r *ApplicationBuilder) WithJobs(fn func() []queue.Job) foundation.ApplicationBuilder {
	r.jobs = fn

	return r
}

func (r *ApplicationBuilder) WithMiddleware(fn func(handler contractsconfiguration.Middleware)) foundation.ApplicationBuilder {
	r.middleware = fn

	return r
}

func (r *ApplicationBuilder) WithMigrations(fn func() []schema.Migration) foundation.ApplicationBuilder {
	r.migrations = fn

	return r
}

func (r *ApplicationBuilder) WithPaths(fn func(paths contractsconfiguration.Paths)) foundation.ApplicationBuilder {
	r.paths = fn

	return r
}

func (r *ApplicationBuilder) WithProviders(fn func() []foundation.ServiceProvider) foundation.ApplicationBuilder {
	r.configuredServiceProviders = fn

	return r
}

func (r *ApplicationBuilder) WithRouting(fn func()) foundation.ApplicationBuilder {
	r.routes = fn

	return r
}

func (r *ApplicationBuilder) WithRules(fn func() []validation.Rule) foundation.ApplicationBuilder {
	r.rules = fn

	return r
}

func (r *ApplicationBuilder) WithSchedule(fn func() []schedule.Event) foundation.ApplicationBuilder {
	r.schedule = fn

	return r
}

func (r *ApplicationBuilder) WithSeeders(fn func() []seeder.Seeder) foundation.ApplicationBuilder {
	r.seeders = fn

	return r
}

func (r *ApplicationBuilder) bootServiceProviders() {
	r.app.BootServiceProviders()
}

func (r *ApplicationBuilder) configureCallback() {
	if r.callback != nil {
		r.callback()
	}
}

func (r *ApplicationBuilder) configureCommands() {
	if r.commands != nil {
		if commands := r.commands(); len(commands) > 0 {
			artisanFacade := r.app.MakeArtisan()
			if artisanFacade == nil {
				color.Errorln("Artisan facade not found, please install it first: ./artisan package:install Artisan")
			} else {
				artisanFacade.Register(commands)
			}
		}
	}
}

func (r *ApplicationBuilder) configureCustomConfig() {
	if r.config != nil {
		r.config()
	}
}

func (r *ApplicationBuilder) configureEventListeners() {
	if r.eventToListeners != nil {
		if eventToListeners := r.eventToListeners(); len(eventToListeners) > 0 {
			eventFacade := r.app.MakeEvent()
			if eventFacade == nil {
				color.Errorln("Event facade not found, please install it first: ./artisan package:install Event")
			} else {
				eventFacade.Register(eventToListeners)
			}
		}
	}
}

func (r *ApplicationBuilder) configureGrpc() {
	var (
		grpcClientInterceptors  map[string][]grpc.UnaryClientInterceptor
		grpcServerInterceptors  []grpc.UnaryServerInterceptor
		grpcClientStatsHandlers map[string][]stats.Handler
		grpcServerStatsHandlers []stats.Handler
	)

	if r.grpcClientInterceptors != nil {
		grpcClientInterceptors = r.grpcClientInterceptors()
	}

	if r.grpcServerInterceptors != nil {
		grpcServerInterceptors = r.grpcServerInterceptors()
	}

	if r.grpcClientStatsHandlers != nil {
		grpcClientStatsHandlers = r.grpcClientStatsHandlers()
	}

	if r.grpcServerStatsHandlers != nil {
		grpcServerStatsHandlers = r.grpcServerStatsHandlers()
	}

	if len(grpcClientInterceptors) > 0 || len(grpcServerInterceptors) > 0 ||
		len(grpcClientStatsHandlers) > 0 || len(grpcServerStatsHandlers) > 0 {
		grpcFacade := r.app.MakeGrpc()
		if grpcFacade == nil {
			color.Errorln("gRPC facade not found, please install it first: ./artisan package:install Grpc")
		} else {
			if len(grpcClientInterceptors) > 0 {
				grpcFacade.UnaryClientInterceptorGroups(grpcClientInterceptors)
			}
			if len(grpcServerInterceptors) > 0 {
				grpcFacade.UnaryServerInterceptors(grpcServerInterceptors)
			}
			if len(grpcClientStatsHandlers) > 0 {
				grpcFacade.ClientStatsHandlerGroups(grpcClientStatsHandlers)
			}
			if len(grpcServerStatsHandlers) > 0 {
				grpcFacade.ServerStatsHandlers(grpcServerStatsHandlers)
			}
		}
	}
}

func (r *ApplicationBuilder) configureJobs() {
	if r.jobs != nil {
		jobs := r.jobs()

		if len(jobs) > 0 {
			queueFacade := r.app.MakeQueue()
			if queueFacade == nil {
				color.Errorln("Queue facade not found, please install it first: ./artisan package:install Queue")
			} else {
				queueFacade.Register(jobs)
			}
		}
	}
}

func (r *ApplicationBuilder) configureMiddleware() {
	if r.middleware != nil {
		routeFacade := r.app.MakeRoute()
		if routeFacade == nil {
			color.Errorln("Route facade not found, please install it first: ./artisan package:install Route")
		} else {
			defaultGlobalMiddleware := routeFacade.GetGlobalMiddleware()
			middleware := configuration.NewMiddleware(defaultGlobalMiddleware)
			r.middleware(middleware)
			routeFacade.SetGlobalMiddleware(middleware.GetGlobalMiddleware())

			if recoveryHandler := middleware.GetRecover(); recoveryHandler != nil {
				routeFacade.Recover(recoveryHandler)
			}
		}
	}
}

func (r *ApplicationBuilder) configureMigrations() {
	if r.migrations != nil {
		if migrations := r.migrations(); len(migrations) > 0 {
			schemaFacade := r.app.MakeSchema()
			if schemaFacade == nil {
				color.Errorln("Schema facade not found, please install it first: ./artisan package:install Schema")
			} else {
				schemaFacade.Register(migrations)
			}
		}
	}
}

func (r *ApplicationBuilder) configurePaths() {
	if r.paths != nil {
		paths := configuration.NewPaths()
		r.paths(paths)
	}
}

func (r *ApplicationBuilder) configureRoutes() {
	if r.routes != nil {
		r.routes()
	}
}

func (r *ApplicationBuilder) configureSchedule() {
	if r.schedule != nil {
		if events := r.schedule(); len(events) > 0 {
			scheduleFacade := r.app.MakeSchedule()
			if scheduleFacade == nil {
				color.Errorln("Schedule facade not found, please install it first: ./artisan package:install Schedule")
			} else {
				scheduleFacade.Register(events)
			}
		}
	}
}

func (r *ApplicationBuilder) configureSeeders() {
	if r.seeders != nil {
		if seeders := r.seeders(); len(seeders) > 0 {
			seederFacade := r.app.MakeSeeder()
			if seederFacade == nil {
				color.Errorln("Seeder facade not found, please install it first: ./artisan package:install Seeder")
			} else {
				seederFacade.Register(seeders)
			}
		}
	}
}

func (r *ApplicationBuilder) configureServiceProviders() {
	if r.configuredServiceProviders != nil {
		configuredServiceProviders := r.configuredServiceProviders()
		if len(configuredServiceProviders) > 0 {
			r.app.AddServiceProviders(configuredServiceProviders)
		}
	}
}

func (r *ApplicationBuilder) configureValidation() {
	var (
		rules   []validation.Rule
		filters []validation.Filter
	)

	if r.rules != nil {
		rules = r.rules()
	}

	if r.filters != nil {
		filters = r.filters()
	}

	if len(rules) > 0 || len(filters) > 0 {
		validationFacade := r.app.MakeValidation()
		if validationFacade == nil {
			color.Errorln("Validation facade not found, please install it first: ./artisan package:install Validation")
		} else {
			if len(rules) > 0 {
				if err := validationFacade.AddRules(rules); err != nil {
					color.Errorf("add validation rules error: %+v", err)
				}
			}
			if len(filters) > 0 {
				if err := validationFacade.AddFilters(filters); err != nil {
					color.Errorf("add validation filters error: %+v", err)
				}
			}
		}
	}
}

func (r *ApplicationBuilder) registerServiceProviders() {
	r.app.RegisterServiceProviders()
}
