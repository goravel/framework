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

var FacadeToPath = map[string]string{
	"Artisan":     "github.com/goravel/framework/console",
	"Auth":        "github.com/goravel/framework/auth",
	"Cache":       "github.com/goravel/framework/cache",
	"Config":      "github.com/goravel/framework/config",
	"Crypt":       "github.com/goravel/framework/crypt",
	"DB":          "github.com/goravel/framework/database",
	"Event":       "github.com/goravel/framework/event",
	"Gate":        "github.com/goravel/framework/auth",
	"Grpc":        "github.com/goravel/framework/grpc",
	"Hash":        "github.com/goravel/framework/hash",
	"Http":        "github.com/goravel/framework/http",
	"Lang":        "github.com/goravel/framework/translation",
	"Log":         "github.com/goravel/framework/log",
	"Mail":        "github.com/goravel/framework/mail",
	"Orm":         "github.com/goravel/framework/database",
	"Queue":       "github.com/goravel/framework/queue",
	"RateLimiter": "github.com/goravel/framework/http",
	"Route":       "github.com/goravel/framework/route",
	"Schedule":    "github.com/goravel/framework/schedule",
	"Schema":      "github.com/goravel/framework/database",
	"Seeder":      "github.com/goravel/framework/database",
	"Session":     "github.com/goravel/framework/session",
	"Storage":     "github.com/goravel/framework/filesystem",
	"Testing":     "github.com/goravel/framework/testing",
	"Validation":  "github.com/goravel/framework/validation",
	"View":        "github.com/goravel/framework/http",
}

type Relationship struct {
	Bindings     []string
	Dependencies []string
	ProvideFor   []string
}
