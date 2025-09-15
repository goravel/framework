package binding

const (
	Artisan     = "goravel.artisan"
	Auth        = "goravel.auth"
	Cache       = "goravel.cache"
	Config      = "goravel.config"
	Crypt       = "goravel.crypt"
	DB          = "goravel.db"
	Event       = "goravel.event"
	Gate        = "goravel.gate"
	Grpc        = "goravel.grpc"
	Hash        = "goravel.hash"
	Http        = "goravel.http"
	Lang        = "goravel.lang"
	Log         = "goravel.log"
	Mail        = "goravel.mail"
	Orm         = "goravel.orm"
	Queue       = "goravel.queue"
	RateLimiter = "goravel.rate_limiter"
	Route       = "goravel.route"
	Schedule    = "goravel.schedule"
	Schema      = "goravel.schema"
	Seeder      = "goravel.seeder"
	Session     = "goravel.session"
	Storage     = "goravel.storage"
	Testing     = "goravel.testing"
	Validation  = "goravel.validation"
	View        = "goravel.view"
)

type Relationship struct {
	Bindings     []string
	Dependencies []string
	ProvideFor   []string
}

type Info struct {
	PkgPath      string
	Dependencies []string
	IsBase       bool
}

var (
	Bindings = map[string]Info{
		Artisan: {
			PkgPath: "github.com/goravel/framework/console",
			Dependencies: []string{
				Config,
			},
			IsBase: true,
		},
		Auth: {
			PkgPath: "github.com/goravel/framework/auth",
			Dependencies: []string{
				Cache,
				Config,
				Log,
				Orm,
			},
		},
		Cache: {
			PkgPath: "github.com/goravel/framework/cache",
			Dependencies: []string{
				Config,
				Log,
			},
		},
		Config: {
			PkgPath: "github.com/goravel/framework/config",
			IsBase:  true,
		},
		Crypt: {
			PkgPath: "github.com/goravel/framework/crypt",
			Dependencies: []string{
				Config,
			},
		},
		DB: {
			PkgPath: "github.com/goravel/framework/database",
			Dependencies: []string{
				Artisan,
				Config,
				Log,
			},
		},
		Event: {
			PkgPath: "github.com/goravel/framework/event",
			Dependencies: []string{
				Queue,
			},
		},
		Gate: {
			PkgPath: "github.com/goravel/framework/auth",
			Dependencies: []string{
				Cache,
				Orm,
			},
		},
		Grpc: {
			PkgPath: "github.com/goravel/framework/grpc",
			Dependencies: []string{
				Config,
			},
		},
		Hash: {
			PkgPath: "github.com/goravel/framework/hash",
			Dependencies: []string{
				Config,
			},
		},
		Http: {
			PkgPath: "github.com/goravel/framework/http",
			Dependencies: []string{
				Cache,
				Config,
				Log,
				Session,
				Validation,
			},
		},
		Lang: {
			PkgPath: "github.com/goravel/framework/translation",
			Dependencies: []string{
				Config,
				Log,
			},
		},
		Log: {
			PkgPath: "github.com/goravel/framework/log",
			Dependencies: []string{
				Config,
			},
		},
		Mail: {
			PkgPath: "github.com/goravel/framework/mail",
			Dependencies: []string{
				Config,
				Queue,
			},
		},
		Orm: {
			PkgPath: "github.com/goravel/framework/database",
			Dependencies: []string{
				Artisan,
				Config,
				Log,
			},
		},
		Queue: {
			PkgPath: "github.com/goravel/framework/queue",
			Dependencies: []string{
				Config,
				DB,
				Log,
			},
		},
		RateLimiter: {
			PkgPath: "github.com/goravel/framework/http",
			Dependencies: []string{
				Cache,
				Config,
				Log,
			},
		},
		Route: {
			PkgPath: "github.com/goravel/framework/route",
			Dependencies: []string{
				Config,
				Http,
			},
		},
		Schedule: {
			PkgPath: "github.com/goravel/framework/schedule",
			Dependencies: []string{
				Artisan,
				Cache,
				Config,
				Log,
			},
		},
		Schema: {
			PkgPath: "github.com/goravel/framework/database",
			Dependencies: []string{
				Config,
				Log,
			},
		},
		Seeder: {
			PkgPath: "github.com/goravel/framework/database",
			Dependencies: []string{
				Artisan,
				Config,
				Log,
			},
		},
		Session: {
			PkgPath: "github.com/goravel/framework/session",
			Dependencies: []string{
				Config,
			},
		},
		Storage: {
			PkgPath: "github.com/goravel/framework/filesystem",
			Dependencies: []string{
				Config,
			},
		},
		Testing: {
			PkgPath: "github.com/goravel/framework/testing",
			Dependencies: []string{
				Artisan,
				Cache,
				Config,
				Orm,
				Route,
				Session,
			},
		},
		Validation: {
			PkgPath: "github.com/goravel/framework/validation",
		},
		View: {
			PkgPath: "github.com/goravel/framework/http",
			Dependencies: []string{
				Cache,
				Config,
				Log,
			},
		},
	}
)
