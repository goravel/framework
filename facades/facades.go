package facades

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
	foundationcontract "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/grpc"
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/process"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/contracts/view"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation"
)

func App() foundationcontract.Application {
	if foundation.App == nil {
		panic(errors.ApplicationNotSet.SetModule(errors.ModuleFacade))
	} else {
		return foundation.App
	}
}

func Artisan() console.Artisan {
	return App().MakeArtisan()
}

func Auth(ctx ...http.Context) auth.Auth {
	return App().MakeAuth(ctx...)
}

func Cache() cache.Cache {
	return App().MakeCache()
}

func Config() config.Config {
	return App().MakeConfig()
}

func Crypt() crypt.Crypt {
	return App().MakeCrypt()
}

func DB() db.DB {
	return App().MakeDB()
}

func Event() event.Instance {
	return App().MakeEvent()
}

func Gate() access.Gate {
	return App().MakeGate()
}

func Grpc() grpc.Grpc {
	return App().MakeGrpc()
}

func Hash() hash.Hash {
	return App().MakeHash()
}

func Http() client.Request {
	return App().MakeHttp()
}

func Lang(ctx context.Context) translation.Translator {
	return App().MakeLang(ctx)
}

func Log() log.Log {
	return App().MakeLog()
}

func Mail() mail.Mail {
	return App().MakeMail()
}

func Orm() orm.Orm {
	return App().MakeOrm()
}

func Process() process.Process {
	return App().MakeProcess()
}

func Queue() queue.Queue {
	return App().MakeQueue()
}

func RateLimiter() http.RateLimiter {
	return App().MakeRateLimiter()
}

func Route() route.Route {
	return App().MakeRoute()
}

func Schedule() schedule.Schedule {
	return App().MakeSchedule()
}

func Schema() schema.Schema {
	return App().MakeSchema()
}

func Seeder() seeder.Facade {
	return App().MakeSeeder()
}

func Session() session.Manager {
	return App().MakeSession()
}

func Storage() filesystem.Storage {
	return App().MakeStorage()
}

func Telemetry() telemetry.Telemetry {
	return App().MakeTelemetry()
}

func Testing() testing.Testing {
	return App().MakeTesting()
}

func Validation() validation.Validation {
	return App().MakeValidation()
}

func View() view.View {
	return App().MakeView()
}
