package modify

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	contractmatch "github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	supportfile "github.com/goravel/framework/support/file"
)

type ModifyActionsTestSuite struct {
	suite.Suite
	config   string
	console  string
	database string
}

func TestModifyActionsTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyActionsTestSuite))
}

func (s *ModifyActionsTestSuite) SetupTest() {
	s.config = `package config

import (	
	"goravel/app/jobs"

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/facades"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
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
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
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
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
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
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
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

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
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

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/facades"
)`)
			},
		},
		{
			name:     "add provider at the beginning",
			content:  s.config,
			matchers: match.Providers(),
			actions: []modify.Action{
				Register("&test.ServiceProvider{}", "*"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&test.ServiceProvider{},
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "add provider (exist)",
			content:  s.config,
			matchers: match.Providers(),
			actions: []modify.Action{
				Register("&crypt.ServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "add provider before",
			content:  s.config,
			matchers: match.Providers(),
			actions: []modify.Action{
				Register("&test.ServiceProvider{}", "&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.Contains(content, `func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":  config.Env("APP_NAME", "Goravel"),
		"exist": map[string]any{},
		"providers": []foundation.ServiceProvider{
			&test.ServiceProvider{},
			&auth.AuthServiceProvider{},
			&crypt.ServiceProvider{},
		},
	})
}`)
			},
		},
		{
			name:     "remove config",
			content:  s.config,
			matchers: match.Config("app"),
			actions: []modify.Action{
				RemoveConfig("providers"),
			},
			assert: func(content string) {
				s.NotContains(content, "providers")
			},
		},
		{
			name:     "remove import(in use)",
			content:  s.config,
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
			content:  s.config,
			matchers: match.Providers(),
			actions: []modify.Action{
				Unregister("&auth.AuthServiceProvider{}"),
			},
			assert: func(content string) {
				s.NotContains(content, "&auth.AuthServiceProvider{}")
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
