package foundation

import (
	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/crypt"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/grpc"
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/contracts/validation"
)

type Container interface {
	// Bind registers a binding with the container.
	Bind(key any, callback func(app Application) (any, error))
	// BindWith registers a binding with the container.
	BindWith(key any, callback func(app Application, parameters map[string]any) (any, error))
	// Instance registers an existing instance as shared in the container.
	Instance(key, instance any)
	// Make resolves the given type from the container.
	Make(key any) (any, error)
	// MakeArtisan resolves the artisan console instance.
	MakeArtisan() console.Artisan
	// MakeAuth resolves the auth instance.
	MakeAuth() auth.Auth
	// MakeCache resolves the cache instance.
	MakeCache() cache.Cache
	// MakeConfig resolves the config instance.
	MakeConfig() config.Config
	// MakeCrypt resolves the crypt instance.
	MakeCrypt() crypt.Crypt
	// MakeEvent resolves the event instance.
	MakeEvent() event.Instance
	// MakeGate resolves the gate instance.
	MakeGate() access.Gate
	// MakeGrpc resolves the grpc instance.
	MakeGrpc() grpc.Grpc
	// MakeHash resolves the hash instance.
	MakeHash() hash.Hash
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
