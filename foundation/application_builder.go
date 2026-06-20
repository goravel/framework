package foundation

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	consoleCommandsFilter      func() []string
	eventToListeners           func() map[event.Event][]event.Listener
	filters                    func() []validation.Filter
	grpcClientCredentials      func() map[string]credentials.TransportCredentials
	grpcClientInterceptors     func() map[string][]grpc.UnaryClientInterceptor
	grpcClientStatsHandlers    func() map[string][]stats.Handler
	grpcServerCredentials      func() credentials.TransportCredentials
	grpcServerInterceptors     func() []grpc.UnaryServerInterceptor
	grpcServerStatsHandlers    func() []stats.Handler
	jobs                       func() []queue.Job
	middleware                 func(middleware contractsconfiguration.Middleware)
	migrations                 func() []schema.Migration
	paths                      func(paths contractsconfiguration.Paths)
	routes                     func()
	rules                      func() []validation.Rule
	runners                    func() []foundation.Runner
	schedule                   func() []schedule.Event
	seeders                    func() []seeder.Seeder
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	return r.app.SetBuilder(r).Build()
}

func (r *ApplicationBuilder) WithCallback(callback func()) foundation.ApplicationBuilder {
	r.callback = callback

	return r
}

func (r *ApplicationBuilder) WithCommands(fn func() []console.Command) foundation.ApplicationBuilder {
	r.commands = fn

	return r
}

// WithCommandsFilter registers a callback that returns the positive list of
// command signatures to keep when the framework registers its own Artisan
// commands. The callback runs once at Build() time; the user can call
// facades.Config() inside it to read app.env or any other setting without
// needing a parameter.
//
// Each entry in the returned slice is matched in one of two ways:
//
//   - Exact match (no wildcard) — checked against command.Signature().
//   - Glob match (the entry contains '*') — checked against
//     command.Signature() using stdpath.Match. '*' matches any sequence
//     of non-'/' characters. '?' is not a wildcard.
//
// Category is never consulted. The filter is signature-only.
//
// Semantics:
//
//   - Method not called           → keep every command (default).
//   - Callback returns nil        → keep every command (no filter).
//   - Callback returns []string{} → drop every command (filter, no matches).
//   - Callback returns entries    → keep only commands whose signature
//     matches an entry (exact or glob).
//
// Composition with WithCommands:
//
// WithCommands adds extra commands to the framework's set; WithCommandsFilter
// then trims the combined set. The filter applies to user-added commands
// too, so the user cannot bypass the filter by adding commands.
func (r *ApplicationBuilder) WithCommandsFilter(fn func() []string) foundation.ApplicationBuilder {
	r.consoleCommandsFilter = fn

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

func (r *ApplicationBuilder) WithGrpcClientCredentials(fn func() map[string]credentials.TransportCredentials) foundation.ApplicationBuilder {
	r.grpcClientCredentials = fn

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

func (r *ApplicationBuilder) WithGrpcServerCredentials(fn func() credentials.TransportCredentials) foundation.ApplicationBuilder {
	r.grpcServerCredentials = fn

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

func (r *ApplicationBuilder) WithRunners(fn func() []foundation.Runner) foundation.ApplicationBuilder {
	r.runners = fn

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
