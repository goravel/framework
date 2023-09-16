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
	// Bind Registers a binding with the container.
	Bind(key any, callback func(app Application) (any, error))
	// BindWith Registers a binding with the container.
	BindWith(key any, callback func(app Application, parameters map[string]any) (any, error))
	// Instance Registers an existing instance as shared in the container.
	Instance(key, instance any)
	// Make Resolves the given type from the container.
	Make(key any) (any, error)
	// MakeArtisan Resolves the artisan console instance.
	MakeArtisan() console.Artisan
	// MakeAuth Resolves the auth instance.
	MakeAuth() auth.Auth
	// MakeCache Resolves the cache instance.
	MakeCache() cache.Cache
	// MakeConfig Resolves the config instance.
	MakeConfig() config.Config
	// MakeCrypt Resolves the crypt instance.
	MakeCrypt() crypt.Crypt
	// MakeEvent Resolves the event instance.
	MakeEvent() event.Instance
	// MakeGate Resolves the gate instance.
	MakeGate() access.Gate
	// MakeGrpc Resolves the grpc instance.
	MakeGrpc() grpc.Grpc
	// MakeHash Resolves the hash instance.
	MakeHash() hash.Hash
	// MakeLog Resolves the log instance.
	MakeLog() log.Log
	// MakeMail Resolves the mail instance.
	MakeMail() mail.Mail
	// MakeOrm Resolves the orm instance.
	MakeOrm() orm.Orm
	// MakeQueue Resolves the queue instance.
	MakeQueue() queue.Queue
	// MakeRateLimiter Resolves the rate limiter instance.
	MakeRateLimiter() http.RateLimiter
	// MakeRoute Resolves the route instance.
	MakeRoute() route.Route
	// MakeSchedule Resolves the schedule instance.
	MakeSchedule() schedule.Schedule
	// MakeStorage Resolves the storage instance.
	MakeStorage() filesystem.Storage
	// MakeTesting Resolves the testing instance.
	MakeTesting() testing.Testing
	// MakeValidation Resolves the validation instance.
	MakeValidation() validation.Validation
	// MakeView Resolves the view instance.
	MakeView() http.View
	// MakeSeeder Resolves the seeder instance.
	MakeSeeder() seeder.Facade
	// MakeWith Resolves the given type with the given parameters from the container.
	MakeWith(key any, parameters map[string]any) (any, error)
	// Singleton Registers a shared binding in the container.
	Singleton(key any, callback func(app Application) (any, error))
}
