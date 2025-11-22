package modify

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	contractmatch "github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
)

type ModifyActionsTestSuite struct {
	suite.Suite
	config    string
	console   string
	database  string
	providers string
}

func TestModifyActionsTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyActionsTestSuite))
}

func (s *ModifyActionsTestSuite) SetupTest() {
	s.config = `package config

import (	
	"goravel/app/jobs"

	"github.com/goravel/framework/facades"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
	})
}`
	s.console = `package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
	"goravel/app/console/commands"
)

type Kernel struct {
}

func (kernel Kernel) Schedule() []schedule.Event {
	return []schedule.Event{}
}

func (kernel Kernel) Commands() []console.Command {
	return []console.Command{}
}`
	s.database = `package database

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/migrations"
	"goravel/database/seeders"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20240915060148CreateUsersTable{},
	}
}

func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
	}
}
`
	s.providers = `package bootstrap

import (
	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&auth.ServiceProvider{},
		&crypt.ServiceProvider{},
	}
}`
}

func (s *ModifyActionsTestSuite) TearDownTest() {}

func (s *ModifyActionsTestSuite) TestActions() {
	tests := []struct {
		name     string
		content  string
		matchers []contractmatch.GoNode
		actions  []modify.Action
		assert   func(filename string)
	}{
		{
			name: "add code to function when function's body is empty",
			content: `package provider
import (
	"github.com/goravel/framework/contracts/foundation"
)
	
type ServiceProvider struct {
}

func (provider *ServiceProvider) Register(app foundation.Application) {}

func (provider *ServiceProvider) Boot(app foundation.Application) {}
`,
			matchers: match.RegisterFunc(),
			actions: []modify.Action{
				Add("facades.Schedule().Register(kernel.Schedule())"),
			},
			assert: func(content string) {
				s.Contains(content, `func (provider *ServiceProvider) Register(app foundation.Application) {
	facades.Schedule().Register(kernel.Schedule())
}`)
			},
		},
		{
			name: "add code to function",
			content: `package provider
import (
	"github.com/goravel/framework/contracts/foundation"
)
	
type ServiceProvider struct {
}

func (provider *ServiceProvider) Register(app foundation.Application) {
	facades.Artisan().Register(kernel.Commands())
}

func (provider *ServiceProvider) Boot(app foundation.Application) {}
`,
			matchers: match.RegisterFunc(),
			actions: []modify.Action{
				Add("facades.Schedule().Register(kernel.Schedule())"),
			},
			assert: func(content string) {
				s.Contains(content, `func (provider *ServiceProvider) Register(app foundation.Application) {
	facades.Artisan().Register(kernel.Commands())
	facades.Schedule().Register(kernel.Schedule())
}`)
			},
		},
		{
			name: "remove code from function",
			content: `package provider
import (
	"github.com/goravel/framework/contracts/foundation"
)
	
type ServiceProvider struct {
}

func (provider *ServiceProvider) Register(app foundation.Application) {
	facades.Artisan().Register(kernel.Commands())
	facades.Schedule().Register(kernel.Schedule())
}

func (provider *ServiceProvider) Boot(app foundation.Application) {}
`,
			matchers: match.RegisterFunc(),
			actions: []modify.Action{
				Remove("facades.Schedule().Register(kernel.Schedule())"),
			},
			assert: func(content string) {
				s.Contains(content, `func (provider *ServiceProvider) Register(app foundation.Application) {
	facades.Artisan().Register(kernel.Commands())
}`)
			},
		},
		{
			name:     "add config (not exist)",
			content:  s.config,
			matchers: match.Config("app"),
			actions: []modify.Action{
				AddConfig("key", `"value"`, "annotation 1", "annotation 2"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		// annotation 1
		// annotation 2
		"key": "value",
	})
}`)
			},
		},
		{
			name:     "add config (exist)",
			content:  s.config,
			matchers: match.Config("app"),
			actions: []modify.Action{
				AddConfig("name", `"Goravel"`),
			},
			assert: func(content string) {
				s.Contains(content, `"name":  "Goravel"`)
			},
		},
		{
			name:     "add config (to map)",
			content:  s.config,
			matchers: match.Config("app.exist"),
			actions: []modify.Action{
				AddConfig("key", `"value"`),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name": config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{
			"key": "value",
		},
	})
}`)
			},
		},
		{
			name:     "add config (with comment)",
			content:  s.config,
			matchers: match.Config("app"),
			actions: []modify.Action{
				AddConfig("drivers", `map[string]any{
    "fiber": map[string]any{
        // prefork mode, see https://docs.gofiber.io/api/fiber/#config
        "prefork": false,
        // Optional, default is 4096 KB
        "body_limit": 4096,
        "header_limit": 4096,
        "route": func() (route.Route, error) {
            return fiberfacades.Route("fiber"), nil
        },
        // Optional, default is "html/template"
        "template": func() (fiber.Views, error) {
            return html.New("./resources/views", ".tmpl"), nil
        },
    },
}`),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"drivers": map[string]any{
			"fiber": map[string]any{
				// prefork mode, see https://docs.gofiber.io/api/fiber/#config
				"prefork": false,
				// Optional, default is 4096 KB
				"body_limit":   4096,
				"header_limit": 4096,
				"route": func() (route.Route, error) {
					return fiberfacades.Route("fiber"), nil
				},
				// Optional, default is "html/template"
				"template": func() (fiber.Views, error) {
					return html.New("./resources/views", ".tmpl"), nil
				},
			},
		},
	})
}`)
			},
		},
		{
			name:     "add import",
			content:  s.config,
			matchers: match.Imports(),
			actions: []modify.Action{
				AddImport("github.com/goravel/test", "t"),
			},
			assert: func(content string) {
				s.Contains(content, `import (
	t "github.com/goravel/test"
	"goravel/app/jobs"

	"github.com/goravel/framework/facades"
)`)
			},
		},
		{
			name:     "add duplicate import",
			content:  s.config,
			matchers: match.Imports(),
			actions: []modify.Action{
				AddImport("goravel/app/jobs"),
			},
			assert: func(content string) {
				s.Contains(content, `import (
	"goravel/app/jobs"

	"github.com/goravel/framework/facades"
)`)
			},
		},
		{
			name:     "add provider at the beginning",
			content:  s.providers,
			matchers: match.Providers(),
			actions: []modify.Action{
				Register("&test.ServiceProvider{}", "*"),
			},
			assert: func(content string) {
				s.Contains(content, `func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&test.ServiceProvider{},
		&auth.ServiceProvider{},
		&crypt.ServiceProvider{},
	}
}`)
			},
		},
		{
			name:     "add provider (exist)",
			content:  s.providers,
			matchers: match.Providers(),
			actions: []modify.Action{
				Register("&crypt.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&auth.ServiceProvider{},
		&crypt.ServiceProvider{},
	}
}`)
			},
		},
		{
			name:     "add provider before",
			content:  s.providers,
			matchers: match.Providers(),
			actions: []modify.Action{
				Register("&test.ServiceProvider{}", "&auth.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&test.ServiceProvider{},
		&auth.ServiceProvider{},
		&crypt.ServiceProvider{},
	}
}`)
			},
		},
		{
			name:     "remove config",
			content:  s.config,
			matchers: match.Config("app"),
			actions: []modify.Action{
				RemoveConfig("exist"),
			},
			assert: func(content string) {
				s.NotContains(content, "exist")
			},
		},
		{
			name:     "remove import(in use)",
			content:  s.providers,
			matchers: match.Imports(),
			actions: []modify.Action{
				RemoveImport("github.com/goravel/framework/auth"),
			},
			assert: func(content string) {
				s.Contains(content, `"github.com/goravel/framework/auth"`)
			},
		},
		{
			name:     "remove import(not in use)",
			content:  strings.Replace(s.config, "&auth.AuthServiceProvider{},", "", 1),
			matchers: match.Imports(),
			actions: []modify.Action{
				RemoveImport("github.com/goravel/framework/auth"),
			},
			assert: func(content string) {
				s.NotContains(content, `"github.com/goravel/framework/auth"`)
			},
		},
		{
			name:     "remove provider",
			content:  s.providers,
			matchers: match.Providers(),
			actions: []modify.Action{
				Unregister("&auth.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.NotContains(content, "&auth.ServiceProvider{}")
			},
		},
		{
			name:     "replace config",
			content:  s.config,
			matchers: match.Config("app"),
			actions: []modify.Action{
				ReplaceConfig("name", `"Goravel"`),
			},
			assert: func(content string) {
				s.Contains(content, `"name":  "Goravel"`)
				s.NotContains(content, `config.Env("APP_NAME", "Goravel")`)
			},
		},
		{
			name:     "add migration",
			content:  s.database,
			matchers: match.Migrations(),
			actions: []modify.Action{
				Register("&migrations.M20250301000000CreateFailedJobsTable{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20240915060148CreateUsersTable{},
		&migrations.M20250301000000CreateFailedJobsTable{},
	}
}`)
			},
		},
		{
			name:     "add seeder",
			content:  s.database,
			matchers: match.Seeders(),
			actions: []modify.Action{
				Register("&seeders.TestSeeder{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
		&seeders.TestSeeder{},
	}
}`)
			},
		},
		{
			name:     "register command",
			content:  s.console,
			matchers: match.Commands(),
			actions: []modify.Action{
				Register("&commands.Test{}"),
			},
			assert: func(content string) {
				s.Contains(content, `
func (kernel Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.Test{},
	}
}`)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.content))
			s.Require().NoError(GoFile(sourceFile).Find(tt.matchers).Modify(tt.actions...).Apply())
			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)
			tt.assert(content)
		})
	}
}

func (s *ModifyActionsTestSuite) TestAddProvider() {
	tests := []struct {
		name              string
		appContent        string
		providersContent  string // empty if file doesn't exist
		pkg               string
		provider          string
		expectedApp       string
		expectedProviders string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add provider when WithProviders doesn't exist and providers.go doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).WithConfig(config.Boot).Run()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
	}
}
`,
		},
		{
			name: "add provider when WithProviders exists with Providers() and providers.go exists",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).WithConfig(config.Boot).Run()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ExistingProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).WithConfig(config.Boot).Run()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ExistingProvider{},
		&providers.AppServiceProvider{},
	}
}
`,
		},
		{
			name: "add provider when WithProviders exists with inline array",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/providers"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders([]foundation.ServiceProvider{
			&providers.ExistingProvider{},
		}).WithConfig(config.Boot).Run()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/providers"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders([]foundation.ServiceProvider{
			&providers.ExistingProvider{},
			&providers.AppServiceProvider{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "error when providers.go exists but WithProviders doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ExistingProvider{},
	}
}
`,
			pkg:               "goravel/app/providers",
			provider:          "&providers.AppServiceProvider{}",
			wantErr:           true,
			expectedErrString: "providers.go already exists but WithProviders is not registered in foundation.Setup()",
		},
		{
			name: "add provider when WithProviders doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.RouteServiceProvider{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).Run()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.RouteServiceProvider{},
	}
}
`,
		},
		{
			name: "add provider from different package",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:      "github.com/goravel/redis",
			provider: "&redis.ServiceProvider{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).WithConfig(config.Boot).Run()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/redis"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&redis.ServiceProvider{},
	}
}
`,
		},
		{
			name: "add multiple providers sequentially",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).WithConfig(config.Boot).Run()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.RouteServiceProvider{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithProviders(Providers()).WithConfig(config.Boot).Run()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
		&providers.RouteServiceProvider{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			providersFile := filepath.Join(bootstrapDir, "providers.go")

			s.Require().NoError(supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				s.NoError(supportfile.Remove(bootstrapDir))
			}()

			if tt.providersContent != "" {
				s.Require().NoError(supportfile.PutContent(providersFile, tt.providersContent))
			}

			err := AddProvider(tt.pkg, tt.provider)

			if tt.wantErr {
				s.Require().Error(err)
				if tt.expectedErrString != "" {
					s.Contains(err.Error(), tt.expectedErrString)
				}
				return
			}

			s.Require().NoError(err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedApp, appContent)

			// Verify providers.go content if expected
			if tt.expectedProviders != "" {
				providersContent, err := supportfile.GetContent(providersFile)
				s.Require().NoError(err)
				s.Equal(tt.expectedProviders, providersContent)
			}
		})
	}
}
