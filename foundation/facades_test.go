package foundation

import (
	"testing"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/database"
	"github.com/stretchr/testify/assert"
)

func TestBindingsToFacades(t *testing.T) {
	assert.ElementsMatch(t, []string{
		"Artisan",
		"Auth",
		"Cache",
		"Config",
		"Crypt",
		"DB",
		"Event",
		"Gate",
		"Grpc",
		"Hash",
		"Http",
		"Lang",
		"Log",
		"Mail",
		"Orm",
		"Queue",
		"RateLimiter",
		"Route",
		"Schedule",
		"Schema",
		"Seeder",
		"Session",
		"Storage",
		"Testing",
		"Validation",
		"View",
	}, bindingsToFacades([]string{
		binding.Artisan,
		binding.Auth,
		binding.Cache,
		binding.Config,
		binding.Crypt,
		binding.DB,
		binding.Event,
		binding.Gate,
		binding.Grpc,
		binding.Hash,
		binding.Http,
		binding.Lang,
		binding.Log,
		binding.Mail,
		binding.Orm,
		binding.Queue,
		binding.RateLimiter,
		binding.Route,
		binding.Schedule,
		binding.Schema,
		binding.Seeder,
		binding.Session,
		binding.Storage,
		binding.Testing,
		binding.Validation,
		binding.View,
	}))
}

func TestGetDependencyBindings(t *testing.T) {
	assert.ElementsMatch(t, []string{
		binding.Cache,
		binding.Config,
		binding.Log,
		binding.Orm,
		binding.Artisan,
	}, getDependencyBindings(binding.Auth))
}

func TestGetFacadeDependencies(t *testing.T) {
	facadeDependencies := getFacadeDependencies()

	assert.ElementsMatch(t, []string{
		"Artisan",
		"Cache",
		"Config",
		"Log",
		"Orm",
	}, facadeDependencies["Auth"])
}

func TestGetFacadePath(t *testing.T) {
	assert.Equal(t, "github.com/goravel/framework/database", getFacadePath(&database.ServiceProvider{}))
}

func TestGetFacadeToPath(t *testing.T) {
	assert.Equal(t, map[string]string{
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
		"Session":     "github.com/goravel/framework/http",
		"Storage":     "github.com/goravel/framework/filesystem",
		"Testing":     "github.com/goravel/framework/testing",
		"Validation":  "github.com/goravel/framework/validation",
		"View":        "github.com/goravel/framework/http",
	}, getFacadeToPath())
}
