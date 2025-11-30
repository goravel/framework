package foundation

import (
	"google.golang.org/grpc"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
)

type ApplicationBuilder interface {
	// Create a new application instance after configuring.
	Create() Application
	// Run the application.
	Run()
	// WithCommands sets the application's commands.
	WithCommands(commands []console.Command) ApplicationBuilder
	// WithConfig sets a callback function to configure the application.
	WithConfig(config func()) ApplicationBuilder
	// WithEvents sets event listeners for the application.
	WithEvents(eventToListeners map[event.Event][]event.Listener) ApplicationBuilder
	// WithFilters sets the application's validation filters.
	WithFilters(filters []validation.Filter) ApplicationBuilder
	// WithGrpcClientInterceptors sets gRPC client interceptors for the application.
	WithGrpcClientInterceptors(groupToInterceptors map[string][]grpc.UnaryClientInterceptor) ApplicationBuilder
	// WithGrpcServerInterceptors sets gRPC server interceptors for the application.
	WithGrpcServerInterceptors(interceptors []grpc.UnaryServerInterceptor) ApplicationBuilder
	// WithJobs registers the application's jobs.
	WithJobs(jobs []queue.Job) ApplicationBuilder
	// WithMiddleware registers the http's middleware.
	WithMiddleware(fn func(handler configuration.Middleware)) ApplicationBuilder
	// WithMigrations registers the database migrations.
	WithMigrations(migrations []schema.Migration) ApplicationBuilder
	// WithPaths sets custom paths for the application.
	WithPaths(fn func(paths configuration.Paths)) ApplicationBuilder
	// WithProviders registers and boots custom service providers.
	WithProviders(providers []ServiceProvider) ApplicationBuilder
	// WithRouting registers the application's routes.
	WithRouting(routes []func()) ApplicationBuilder
	// WithRules registers the custom validation rules.
	WithRules(rules []validation.Rule) ApplicationBuilder
	// WithSchedule sets scheduled events for the application.
	WithSchedule(events []schedule.Event) ApplicationBuilder
	// WithSeeders registers the database seeders.
	WithSeeders(seeders []seeder.Seeder) ApplicationBuilder
}
