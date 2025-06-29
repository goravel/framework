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
