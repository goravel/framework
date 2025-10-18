package foundation

import (
	"context"

	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/crypt"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/grpc"
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/contracts/validation"
)

type Runner interface {
	ShouldRun() bool
	Run() error
	Shutdown() error
}

type AboutItem struct {
	Key   string
	Value string
}

type ApplicationBuilder interface {
	// Create a new application instance after configuring.
	Create() Application
	// Run the application.
	Run()
	// WithConfig sets a callback function to configure the application.
	WithConfig(func()) ApplicationBuilder
	// WithEvents sets event listeners for the application.
	WithEvents(map[event.Event][]event.Listener) ApplicationBuilder
}

type Application interface {
	// About add information to the application's about command.
	About(section string, items []AboutItem)
	// Boot register and bootstrap configured service providers.
	Boot()
	// Commands register the given commands with the console application.
	Commands([]console.Command)
	// Context gets the application context.
	Context() context.Context
	// GetJson get the JSON implementation.
	GetJson() Json
	// IsLocale get the current application locale.
	IsLocale(ctx context.Context, locale string) bool
	// Publishes register the given paths to be published by the "vendor:publish" command.
	Publishes(packageName string, paths map[string]string, groups ...string)
	// Refresh all modules after changing config, will call the Boot method simultaneously.
	Refresh()
	// Run runs modules.
	Run(runners ...Runner)
	// SetJson set the JSON implementation.
	SetJson(json Json)
	// SetLocale set the current application locale.
	SetLocale(ctx context.Context, locale string) context.Context
	// Shutdown the application and all its runners.
	Shutdown()
	// Version gets the version number of the application.
	Version() string

	// Paths
	// BasePath get the base path of the Goravel installation.
	BasePath(path ...string) string
	// ConfigPath get the path to the configuration files.
	ConfigPath(path ...string) string
	// CurrentLocale get the current application locale.
	CurrentLocale(ctx context.Context) string
	// DatabasePath get the path to the database directory.
	DatabasePath(path ...string) string
	// ExecutablePath get the path to the executable of the running Goravel application.
	ExecutablePath(path ...string) string
	// FacadePath get the path to the facade files.
	FacadesPath(path ...string) string
	// LangPath get the path to the language files.
	LangPath(path ...string) string
	// Path gets the path respective to "app" directory.
	Path(path ...string) string
	// PublicPath get the path to the public directory.
	PublicPath(path ...string) string
	// ResourcePath get the path to the resources directory.
	ResourcePath(path ...string) string
	// StoragePath get the path to the storage directory.
	StoragePath(path ...string) string

	// Container
	// Bind registers a binding with the container.
	Bind(key any, callback func(app Application) (any, error))
	// Bindings returns all bindings registered in the container.
	Bindings() []any
	// BindWith registers a binding with the container.
	BindWith(key any, callback func(app Application, parameters map[string]any) (any, error))
	// Fresh modules after changing config, will fresh all bindings except for config if no bindings provided.
	// Notice, this method only freshs the facade, if another facade injects the facade previously, the another
	// facades should be fresh simulaneously.
	Fresh(bindings ...any)
	// Instance registers an existing instance as shared in the container.
	Instance(key, instance any)
	// Make resolves the given type from the container.
	Make(key any) (any, error)
	// MakeArtisan resolves the artisan console instance.
	MakeArtisan() console.Artisan
	// MakeAuth resolves the auth instance.
	MakeAuth(ctx ...http.Context) auth.Auth
	// MakeCache resolves the cache instance.
	MakeCache() cache.Cache
	// MakeConfig resolves the config instance.
	MakeConfig() config.Config
	// MakeCrypt resolves the crypt instance.
	MakeCrypt() crypt.Crypt
	// MakeDB resolves the db instance.
	MakeDB() db.DB
	// MakeEvent resolves the event instance.
	MakeEvent() event.Instance
	// MakeGate resolves the gate instance.
	MakeGate() access.Gate
	// MakeGrpc resolves the grpc instance.
	MakeGrpc() grpc.Grpc
	// MakeHash resolves the hash instance.
	MakeHash() hash.Hash
	// MakeHttp resolves the http instance.
	MakeHttp() client.Request
	// MakeLang resolves the lang instance.
	MakeLang(ctx context.Context) translation.Translator
	// MakeLog resolves the log instance.
	MakeLog() log.Log
	// MakeMail resolves the mail instance.
	MakeMail() mail.Mail
	// MakeOrm resolves the orm instance.
	MakeOrm() orm.Orm
	// MakeQueue resolves the queue instance.
	MakeQueue() queue.Queue
	// MakeRateLimiter resolves the rate limiter instance.
	MakeRateLimiter() http.RateLimiter
	// MakeRoute resolves the route instance.
	MakeRoute() route.Route
	// MakeSchedule resolves the schedule instance.
	MakeSchedule() schedule.Schedule
	// MakeSchema resolves the schema instance.
	MakeSchema() schema.Schema
	// MakeSession resolves the session instance.
	MakeSession() session.Manager
	// MakeStorage resolves the storage instance.
	MakeStorage() filesystem.Storage
	// MakeTesting resolves the testing instance.
	MakeTesting() testing.Testing
	// MakeValidation resolves the validation instance.
	MakeValidation() validation.Validation
	// MakeView resolves the view instance.
	MakeView() http.View
	// MakeSeeder resolves the seeder instance.
	MakeSeeder() seeder.Facade
	// MakeWith resolves the given type with the given parameters from the container.
	MakeWith(key any, parameters map[string]any) (any, error)
	// Singleton registers a shared binding in the container.
	Singleton(key any, callback func(app Application) (any, error))
}
