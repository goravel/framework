package foundation

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
)

var (
	facades = map[string]foundation.FacadeInfo{
		"Artisan": {
			Binding:      binding.Artisan,
			PkgPath:      "github.com/goravel/framework/console",
			Dependencies: []string{"Config"},
			IsBase:       true,
		},
		"Auth": {
			Binding:      binding.Auth,
			PkgPath:      "github.com/goravel/framework/auth",
			Dependencies: []string{"Cache", "Config", "Log", "Orm"},
		},
		"Cache": {
			Binding:      binding.Cache,
			PkgPath:      "github.com/goravel/framework/cache",
			Dependencies: []string{"Config", "Log"},
		},
		"Config": {
			Binding: binding.Config,
			PkgPath: "github.com/goravel/framework/config",
			IsBase:  true,
		},
		"Crypt": {
			Binding:      binding.Crypt,
			PkgPath:      "github.com/goravel/framework/crypt",
			Dependencies: []string{"Config"},
		},
		"DB": {
			Binding:      binding.DB,
			PkgPath:      "github.com/goravel/framework/database",
			Dependencies: []string{"Artisan", "Config", "Log"},
		},
		"Event": {
			Binding:      binding.Event,
			PkgPath:      "github.com/goravel/framework/event",
			Dependencies: []string{"Queue"},
		},
		"Gate": {
			Binding:      binding.Gate,
			PkgPath:      "github.com/goravel/framework/auth",
			Dependencies: []string{"Cache", "Config", "Log", "Orm"},
		},
		"Grpc": {
			Binding:      binding.Grpc,
			PkgPath:      "github.com/goravel/framework/grpc",
			Dependencies: []string{"Config"},
		},
		"Hash": {
			Binding:      binding.Hash,
			PkgPath:      "github.com/goravel/framework/hash",
			Dependencies: []string{"Config"},
		},
		"Http": {
			Binding:      binding.Http,
			PkgPath:      "github.com/goravel/framework/http",
			Dependencies: []string{"Cache", "Config", "Log"},
		},
		"Lang": {
			Binding:      binding.Lang,
			PkgPath:      "github.com/goravel/framework/translation",
			Dependencies: []string{"Config", "Log"},
		},
		"Log": {
			Binding:      binding.Log,
			PkgPath:      "github.com/goravel/framework/log",
			Dependencies: []string{"Config"},
		},
		"Mail": {
			Binding:      binding.Mail,
			PkgPath:      "github.com/goravel/framework/mail",
			Dependencies: []string{"Config", "Queue"},
		},
		"Orm": {
			Binding:      binding.Orm,
			PkgPath:      "github.com/goravel/framework/database",
			Dependencies: []string{"Artisan", "Config", "Log"},
		},
		"Queue": {
			Binding:      binding.Queue,
			PkgPath:      "github.com/goravel/framework/queue",
			Dependencies: []string{"Config", "DB", "Log"},
		},
		"RateLimiter": {
			Binding:      binding.RateLimiter,
			PkgPath:      "github.com/goravel/framework/http",
			Dependencies: []string{"Cache", "Config", "Log"},
		},
		"Route": {
			Binding:      binding.Route,
			PkgPath:      "github.com/goravel/framework/route",
			Dependencies: []string{"Config"},
		},
		"Schedule": {
			Binding:      binding.Schedule,
			PkgPath:      "github.com/goravel/framework/schedule",
			Dependencies: []string{"Artisan", "Cache", "Config", "Log"},
		},
		"Schema": {
			Binding:      binding.Schema,
			PkgPath:      "github.com/goravel/framework/database",
			Dependencies: []string{"Artisan", "Config", "Log"},
		},
		"Seeder": {
			Binding:      binding.Seeder,
			PkgPath:      "github.com/goravel/framework/database",
			Dependencies: []string{"Artisan", "Config", "Log"},
		},
		"Session": {
			Binding:      binding.Session,
			PkgPath:      "github.com/goravel/framework/http",
			Dependencies: []string{"Cache", "Config", "Log"},
		},
		"Storage": {
			Binding:      binding.Storage,
			PkgPath:      "github.com/goravel/framework/filesystem",
			Dependencies: []string{"Config"},
		},
		"Testing": {
			Binding:      binding.Testing,
			PkgPath:      "github.com/goravel/framework/testing",
			Dependencies: []string{"Artisan", "Cache", "Config", "Orm"},
		},
		"Validation": {
			Binding: binding.Validation,
			PkgPath: "github.com/goravel/framework/validation",
		},
		"View": {
			Binding:      binding.View,
			PkgPath:      "github.com/goravel/framework/http",
			Dependencies: []string{"Cache", "Config", "Log"},
		},
	}
)
