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
	"github.com/goravel/framework/contracts/validation"
)

type Container interface {
	Bind(key any, callback func(app Application) (any, error))
	BindWith(key any, callback func(app Application, parameters map[string]any) (any, error))
	Instance(key, instance any)
	Make(key any) (any, error)
	MakeArtisan() console.Artisan
	MakeAuth() auth.Auth
	MakeCache() cache.Cache
	MakeConfig() config.Config
	MakeCrypt() crypt.Crypt
	MakeEvent() event.Instance
	MakeGate() access.Gate
	MakeGrpc() grpc.Grpc
	MakeHash() hash.Hash
	MakeLog() log.Log
	MakeMail() mail.Mail
	MakeOrm() orm.Orm
	MakeQueue() queue.Queue
	MakeRateLimiter() http.RateLimiter
	MakeRoute() route.Engine
	MakeSchedule() schedule.Schedule
	MakeStorage() filesystem.Storage
	MakeValidation() validation.Validation
	MakeSeeder() seeder.Facade
	MakeWith(key any, parameters map[string]any) (any, error)
	Singleton(key any, callback func(app Application) (any, error))
}
