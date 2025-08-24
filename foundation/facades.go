package foundation

import (
	"reflect"

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/cache"
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/database"
	"github.com/goravel/framework/event"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/grpc"
	"github.com/goravel/framework/hash"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/log"
	"github.com/goravel/framework/mail"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/schedule"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/testing"
	"github.com/goravel/framework/translation"
	"github.com/goravel/framework/validation"
)

type facadeInfo struct {
	binding         string
	serviceProvider foundation.ServiceProvider
}

var facades = map[string]facadeInfo{
	"Artisan": {
		binding:         binding.Artisan,
		serviceProvider: &console.ServiceProvider{},
	},
	"Auth": {
		binding:         binding.Auth,
		serviceProvider: &auth.ServiceProvider{},
	},
	"Cache": {
		binding:         binding.Cache,
		serviceProvider: &cache.ServiceProvider{},
	},
	"Config": {
		binding:         binding.Config,
		serviceProvider: &config.ServiceProvider{},
	},
	"Crypt": {
		binding:         binding.Crypt,
		serviceProvider: &crypt.ServiceProvider{},
	},
	"DB": {
		binding:         binding.DB,
		serviceProvider: &database.ServiceProvider{},
	},
	"Event": {
		binding:         binding.Event,
		serviceProvider: &event.ServiceProvider{},
	},
	"Gate": {
		binding:         binding.Gate,
		serviceProvider: &auth.ServiceProvider{},
	},
	"Grpc": {
		binding:         binding.Grpc,
		serviceProvider: &grpc.ServiceProvider{},
	},
	"Hash": {
		binding:         binding.Hash,
		serviceProvider: &hash.ServiceProvider{},
	},
	"Http": {
		binding:         binding.Http,
		serviceProvider: &http.ServiceProvider{},
	},
	"Lang": {
		binding:         binding.Lang,
		serviceProvider: &translation.ServiceProvider{},
	},
	"Log": {
		binding:         binding.Log,
		serviceProvider: &log.ServiceProvider{},
	},
	"Mail": {
		binding:         binding.Mail,
		serviceProvider: &mail.ServiceProvider{},
	},
	"Orm": {
		binding:         binding.Orm,
		serviceProvider: &database.ServiceProvider{},
	},
	"Queue": {
		binding:         binding.Queue,
		serviceProvider: &queue.ServiceProvider{},
	},
	"RateLimiter": {
		binding:         binding.RateLimiter,
		serviceProvider: &http.ServiceProvider{},
	},
	"Route": {
		binding:         binding.Route,
		serviceProvider: &http.ServiceProvider{},
	},
	"Schedule": {
		binding:         binding.Schedule,
		serviceProvider: &schedule.ServiceProvider{},
	},
	"Schema": {
		binding:         binding.Schema,
		serviceProvider: &database.ServiceProvider{},
	},
	"Seeder": {
		binding:         binding.Seeder,
		serviceProvider: &database.ServiceProvider{},
	},
	"Session": {
		binding:         binding.Session,
		serviceProvider: &http.ServiceProvider{},
	},
	"Storage": {
		binding:         binding.Storage,
		serviceProvider: &filesystem.ServiceProvider{},
	},
	"Testing": {
		binding:         binding.Testing,
		serviceProvider: &testing.ServiceProvider{},
	},
	"Validation": {
		binding:         binding.Validation,
		serviceProvider: &validation.ServiceProvider{},
	},
	"View": {
		binding:         binding.View,
		serviceProvider: &http.ServiceProvider{},
	},
}

func getFacadeDependencies() map[string][]string {
	dependencies := make(map[string][]string)

	for facade, info := range facades {
		dependencyBindings := getDependencyBindings(info.binding)
		dependencyFacades := bindingsToFacades(dependencyBindings)

		dependencies[facade] = dependencyFacades
	}

	return dependencies
}

func getFacadePath(serviceProvider foundation.ServiceProvider) string {
	t := reflect.TypeOf(serviceProvider)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.PkgPath()
}

func getFacadeToPath() map[string]string {
	facadeToPath := make(map[string]string)

	for facade, info := range facades {
		facadeToPath[facade] = getFacadePath(info.serviceProvider)
	}

	return facadeToPath
}

func getDependencyBindings(binding string) []string {
	for _, info := range facades {
		if info.binding == binding {
			serviceProviderWithRelations, ok := info.serviceProvider.(foundation.ServiceProviderWithRelations)
			if !ok {
				continue
			}

			dependencyBindings := serviceProviderWithRelations.Relationship().Dependencies
			if len(dependencyBindings) == 0 {
				return nil
			}

			allDependencyBindings := make([]string, len(dependencyBindings))
			copy(allDependencyBindings, dependencyBindings)

			for _, dependencyBinding := range dependencyBindings {
				subDependencyBindings := getDependencyBindings(dependencyBinding)
				allDependencyBindings = append(allDependencyBindings, subDependencyBindings...)
			}

			return collect.Unique(allDependencyBindings)
		}
	}

	return nil
}

func bindingsToFacades(bindings []string) []string {
	result := make([]string, 0)

	for _, binding := range bindings {
		for facade, info := range facades {
			if info.binding == binding {
				result = append(result, facade)
				break
			}
		}
	}

	return result
}
