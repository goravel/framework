package foundation

import (
	"testing"

	"github.com/goravel/framework/contracts/binding"
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

func TestGetDependencyBindings(t *testing.T) {
	assert.ElementsMatch(t, []string{
		binding.Cache,
		binding.Config,
		binding.Log,
		binding.Orm,
		binding.Artisan,
	}, getDependencyBindings(binding.Auth))
}
