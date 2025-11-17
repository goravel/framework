package modify

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/suite"

	contractmatch "github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
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

func (s *ModifyActionsTestSuite) TestAddMiddleware() {
	tests := []struct {
		name     string
		content  string
		pkg      string
		mw       string
		expected string
		wantErr  bool
	}{
		{
			name: "add middleware when WithMiddleware doesn't exist",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Auth{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add middleware when WithMiddleware already exists",
			content: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{})
		}).WithConfig(config.Boot).Run()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Auth{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{},
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add middleware with complex chain",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Cors{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Cors{},
			)
		}).WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
		},
		{
			name: "add middleware when WithMiddleware exists but no Append call",
			content: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).WithConfig(config.Boot).Run()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Auth{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add middleware to Boot function with multiple statements",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	foundation.Setup().WithConfig(config.Boot).Run()
	app.Start()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Throttle{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Throttle{},
			)
		}).WithConfig(config.Boot).Run()
	app.Start()
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.content))

			// Override the Config.Paths.App for testing
			originalAppPath := support.Config.Paths.App
			support.Config.Paths.App = sourceFile
			defer func() {
				support.Config.Paths.App = originalAppPath
			}()

			err := AddMiddleware(tt.pkg, tt.mw)
			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)
			s.Equal(tt.expected, content)
		})
	}
}

func (s *ModifyActionsTestSuite) Test_appendToExistingMiddleware() {
	tests := []struct {
		name              string
		initialContent    string
		middlewareToAdd   string
		expectedArgsCount int
	}{
		{
			name: "append to existing Append call",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Auth{})
		}).Run()
}`,
			middlewareToAdd:   "&middleware.Cors{}",
			expectedArgsCount: 2,
		},
		{
			name: "append to empty function",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).Run()
}`,
			middlewareToAdd:   "&middleware.Auth{}",
			expectedArgsCount: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the WithMiddleware call
			var withMiddlewareCall *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if sel.Sel.Name == "WithMiddleware" {
							withMiddlewareCall = call
							return false
						}
					}
				}
				return true
			})

			s.NotNil(withMiddlewareCall, "Expected to find WithMiddleware call")

			middlewareExpr := MustParseExpr(tt.middlewareToAdd).(dst.Expr)
			appendToExistingMiddleware(withMiddlewareCall, middlewareExpr)

			funcLit := withMiddlewareCall.Args[0].(*dst.FuncLit)
			appendCall := findMiddlewareAppendCall(funcLit)

			s.NotNil(appendCall, "Expected Append call to exist after modification")
			s.Equal(tt.expectedArgsCount, len(appendCall.Args), "Expected %d arguments in Append call", tt.expectedArgsCount)
		})
	}
}

func (s *ModifyActionsTestSuite) Test_addMiddlewareAppendCall() {
	tests := []struct {
		name            string
		initialContent  string
		middlewareToAdd string
	}{
		{
			name: "add Append to empty function",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).Run()
}`,
			middlewareToAdd: "&middleware.Auth{}",
		},
		{
			name: "add Append to function with other statements",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
		handler.Register(&middleware.Other{})
	}).Run()
}`,
			middlewareToAdd: "&middleware.Cors{}",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the function literal
			var funcLit *dst.FuncLit
			dst.Inspect(file, func(n dst.Node) bool {
				if fl, ok := n.(*dst.FuncLit); ok {
					funcLit = fl
					return false
				}
				return true
			})

			s.NotNil(funcLit, "Expected to find function literal")

			originalStmtCount := len(funcLit.Body.List)
			middlewareExpr := MustParseExpr(tt.middlewareToAdd).(dst.Expr)

			addMiddlewareAppendCall(funcLit, middlewareExpr)

			s.Equal(originalStmtCount+1, len(funcLit.Body.List), "Expected one more statement")

			appendCall := findMiddlewareAppendCall(funcLit)
			s.NotNil(appendCall, "Expected to find newly added Append call")
			s.Equal(1, len(appendCall.Args), "Expected exactly 1 argument in Append call")
		})
	}
}

func (s *ModifyActionsTestSuite) Test_addMiddlewareImports() {
	tests := []struct {
		name             string
		initialContent   string
		pkg              string
		expectError      bool
		expectedImports  []string
		unexpectedImport string
	}{
		{
			name: "add middleware imports to file with existing imports",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "github.com/goravel/framework/http/middleware",
			expectError: false,
			expectedImports: []string{
				"github.com/goravel/framework/http/middleware",
				"github.com/goravel/framework/contracts/foundation/configuration",
			},
		},
		{
			name: "add middleware imports when configuration import already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "github.com/goravel/framework/http/middleware",
			expectError: false,
			expectedImports: []string{
				"github.com/goravel/framework/http/middleware",
				"github.com/goravel/framework/contracts/foundation/configuration",
			},
		},
		{
			name: "add middleware imports when middleware package already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "github.com/goravel/framework/http/middleware",
			expectError: false,
			expectedImports: []string{
				"github.com/goravel/framework/http/middleware",
				"github.com/goravel/framework/contracts/foundation/configuration",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			err := addMiddlewareImports(sourceFile, tt.pkg)

			if tt.expectError {
				s.Error(err)
				return
			}

			s.NoError(err)

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			for _, expectedImport := range tt.expectedImports {
				s.Contains(content, expectedImport, "Expected import %s to be present", expectedImport)
			}

			if tt.unexpectedImport != "" {
				s.NotContains(content, tt.unexpectedImport)
			}
		})
	}
}

func (s *ModifyActionsTestSuite) Test_createWithMiddleware() {
	tests := []struct {
		name            string
		initialContent  string
		middlewareToAdd string
		expectedContent string
	}{
		{
			name: "create WithMiddleware and insert into chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedContent: `WithMiddleware(func(handler configuration.Middleware) {
		handler.Append(
			&middleware.Auth{},
		)
	}).WithConfig(config.Boot)`,
		},
		{
			name: "create WithMiddleware when multiple chain calls exist",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
			middlewareToAdd: "&middleware.Cors{}",
			expectedContent: `WithMiddleware(func(handler configuration.Middleware) {
		handler.Append(
			&middleware.Cors{},
		)
	}).WithConfig(config.Boot)`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the foundation.Setup() call and the chain
			var setupCall *dst.CallExpr
			var parentOfSetup *dst.SelectorExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if innerCall, ok := sel.X.(*dst.CallExpr); ok {
							if innerSel, ok := innerCall.Fun.(*dst.SelectorExpr); ok {
								if innerSel.Sel.Name == "Setup" {
									setupCall = innerCall
									parentOfSetup = sel
									return false
								}
							}
						}
					}
				}
				return true
			})

			s.NotNil(setupCall, "Expected to find Setup call")
			s.NotNil(parentOfSetup, "Expected to find parent of Setup")

			middlewareExpr := MustParseExpr(tt.middlewareToAdd).(dst.Expr)
			createWithMiddleware(setupCall, parentOfSetup, middlewareExpr)

			// Verify the structure was created
			s.NotNil(parentOfSetup.X, "Expected parentOfSetup.X to be updated")
			withMiddlewareCall, ok := parentOfSetup.X.(*dst.CallExpr)
			s.True(ok, "Expected parentOfSetup.X to be a CallExpr")

			sel, ok := withMiddlewareCall.Fun.(*dst.SelectorExpr)
			s.True(ok, "Expected WithMiddleware fun to be a SelectorExpr")
			s.Equal("WithMiddleware", sel.Sel.Name)

			s.Require().Len(withMiddlewareCall.Args, 1)
			funcLit, ok := withMiddlewareCall.Args[0].(*dst.FuncLit)
			s.True(ok, "Expected first argument to be a function literal")

			appendCall := findMiddlewareAppendCall(funcLit)
			s.NotNil(appendCall, "Expected Append call to exist")
			s.Equal(1, len(appendCall.Args), "Expected exactly 1 argument in Append call")
		})
	}
}

func (s *ModifyActionsTestSuite) Test_containsFoundationSetup() {
	tests := []struct {
		name     string
		stmt     string
		expected bool
	}{
		{
			name:     "contains foundation.Setup()",
			stmt:     `foundation.Setup().Run()`,
			expected: true,
		},
		{
			name:     "contains foundation.Setup() in chain",
			stmt:     `foundation.Setup().WithConfig(config.Boot).Run()`,
			expected: true,
		},
		{
			name:     "does not contain foundation.Setup()",
			stmt:     `app.Run()`,
			expected: false,
		},
		{
			name:     "contains Setup() but not from foundation",
			stmt:     `something.Setup().Run()`,
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expr := MustParseExpr(tt.stmt).(dst.Expr)
			stmt := &dst.ExprStmt{X: expr}

			result := containsFoundationSetup(stmt)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ModifyActionsTestSuite) Test_findFoundationSetupCallsForMiddleware() {
	tests := []struct {
		name                    string
		initialContent          string
		expectSetup             bool
		expectWithMiddleware    bool
		expectParentOfSetup     bool
		withMiddlewareArgsCount int
	}{
		{
			name: "find Setup without WithMiddleware",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			expectSetup:          true,
			expectWithMiddleware: false,
			expectParentOfSetup:  true,
		},
		{
			name: "find Setup with WithMiddleware",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Auth{})
		}).WithConfig(config.Boot).Run()
}
`,
			expectSetup:             true,
			expectWithMiddleware:    true,
			expectParentOfSetup:     true,
			withMiddlewareArgsCount: 1,
		},
		{
			name: "find Setup with complex chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).WithSchedule(schedule.Boot).Run()
}
`,
			expectSetup:          true,
			expectWithMiddleware: false,
			expectParentOfSetup:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the main call expression
			var mainCallExpr *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if stmt, ok := n.(*dst.ExprStmt); ok {
					if call, ok := stmt.X.(*dst.CallExpr); ok {
						mainCallExpr = call
						return false
					}
				}
				return true
			})

			s.NotNil(mainCallExpr, "Expected to find main call expression")

			setupCall, withMiddlewareCall, parentOfSetup := findFoundationSetupCallsForMiddleware(mainCallExpr)

			if tt.expectSetup {
				s.NotNil(setupCall, "Expected to find Setup call")
				sel, ok := setupCall.Fun.(*dst.SelectorExpr)
				s.True(ok)
				s.Equal("Setup", sel.Sel.Name)
			} else {
				s.Nil(setupCall, "Expected not to find Setup call")
			}

			if tt.expectWithMiddleware {
				s.NotNil(withMiddlewareCall, "Expected to find WithMiddleware call")
				sel, ok := withMiddlewareCall.Fun.(*dst.SelectorExpr)
				s.True(ok)
				s.Equal("WithMiddleware", sel.Sel.Name)
				s.Equal(tt.withMiddlewareArgsCount, len(withMiddlewareCall.Args))
			} else {
				s.Nil(withMiddlewareCall, "Expected not to find WithMiddleware call")
			}

			if tt.expectParentOfSetup {
				s.NotNil(parentOfSetup, "Expected to find parent of Setup")
			} else {
				s.Nil(parentOfSetup, "Expected not to find parent of Setup")
			}
		})
	}
}

func (s *ModifyActionsTestSuite) Test_findMiddlewareAppendCall() {
	tests := []struct {
		name           string
		initialContent string
		expectFound    bool
		expectedArgs   int
	}{
		{
			name: "find Append call with single argument",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().WithMiddleware(func(handler configuration.Middleware) {
		handler.Append(&middleware.Auth{})
	}).Run()
}`,
			expectFound:  true,
			expectedArgs: 1,
		},
		{
			name: "find Append call with multiple arguments",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Auth{}, &middleware.Cors{})
		}).Run()
}`,
			expectFound:  true,
			expectedArgs: 2,
		},
		{
			name: "return nil when no Append call exists",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).Run()
}`,
			expectFound: false,
		},
		{
			name: "return nil when function has other calls but not Append",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Register(&middleware.Auth{})
		}).Run()
}`,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the function literal
			var funcLit *dst.FuncLit
			dst.Inspect(file, func(n dst.Node) bool {
				if fl, ok := n.(*dst.FuncLit); ok {
					funcLit = fl
					return false
				}
				return true
			})

			s.NotNil(funcLit, "Expected to find function literal")

			appendCall := findMiddlewareAppendCall(funcLit)

			if tt.expectFound {
				s.NotNil(appendCall, "Expected to find Append call")
				s.Equal(tt.expectedArgs, len(appendCall.Args), "Expected %d arguments in Append call", tt.expectedArgs)
			} else {
				s.Nil(appendCall, "Expected not to find Append call")
			}
		})
	}
}

func (s *ModifyActionsTestSuite) Test_foundationSetupMiddleware() {
	tests := []struct {
		name            string
		initialContent  string
		middlewareToAdd string
		expectedResult  string
	}{
		{
			name: "modify chain without WithMiddleware",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "modify chain with existing WithMiddleware",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{})
		}).WithConfig(config.Boot).Run()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{},
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "skip non-foundation.Setup() statements",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
			middlewareToAdd: "&middleware.Auth{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			_, err = decorator.Parse(content)
			s.Require().NoError(err)

			// Apply the action
			err = GoFile(sourceFile).Find(match.FoundationSetup()).Modify(foundationSetupMiddleware(tt.middlewareToAdd)).Apply()
			s.NoError(err)

			// Read the result
			resultContent, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			s.Equal(tt.expectedResult, resultContent)
		})
	}
}

func (s *ModifyActionsTestSuite) TestAddSeeder() {
	tests := []struct {
		name     string
		content  string
		pkg      string
		seeder   string
		expected string
		wantErr  bool
	}{
		{
			name: "add seeder when WithSeeders doesn't exist",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.DatabaseSeeder{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.DatabaseSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add seeder when WithSeeders already exists",
			content: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.ExistingSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.DatabaseSeeder{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.ExistingSeeder{},
			&seeders.DatabaseSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add seeder with complex chain",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.UserSeeder{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.UserSeeder{},
		}).WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
		},
		{
			name: "add seeder to Boot function with multiple statements",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	foundation.Setup().WithConfig(config.Boot).Run()
	app.Start()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.ProductSeeder{}",
			expected: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	app := foundation.NewApplication()
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.ProductSeeder{},
		}).WithConfig(config.Boot).Run()
	app.Start()
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.content))

			// Override the Config.Paths.App for testing
			originalAppPath := support.Config.Paths.App
			support.Config.Paths.App = sourceFile
			defer func() {
				support.Config.Paths.App = originalAppPath
			}()

			err := AddSeeder(tt.pkg, tt.seeder)
			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)
			s.Equal(tt.expected, content)
		})
	}
}

func (s *ModifyActionsTestSuite) Test_appendToExistingSeeder() {
	tests := []struct {
		name              string
		initialContent    string
		seederToAdd       string
		expectedArgsCount int
	}{
		{
			name: "append to existing WithSeeders call",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.ExistingSeeder{},
		}).Run()
}`,
			seederToAdd:       "&seeders.DatabaseSeeder{}",
			expectedArgsCount: 2,
		},
		{
			name: "append to empty seeder array",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{}).Run()
}`,
			seederToAdd:       "&seeders.DatabaseSeeder{}",
			expectedArgsCount: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			file, err := decorator.Parse(content)
			s.Require().NoError(err)

			// Find the WithSeeders call
			var withSeedersCall *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if sel.Sel.Name == "WithSeeders" {
							withSeedersCall = call
							return false
						}
					}
				}
				return true
			})

			s.Require().NotNil(withSeedersCall, "WithSeeders call not found")

			seederExpr := MustParseExpr(tt.seederToAdd).(dst.Expr)
			appendToExistingSeeder(withSeedersCall, seederExpr)

			// Verify the seeder was appended
			s.Require().Len(withSeedersCall.Args, 1)
			compositeLit, ok := withSeedersCall.Args[0].(*dst.CompositeLit)
			s.Require().True(ok)
			s.Equal(tt.expectedArgsCount, len(compositeLit.Elts))
		})
	}
}

func (s *ModifyActionsTestSuite) Test_addSeederImports() {
	tests := []struct {
		name             string
		initialContent   string
		pkg              string
		expectError      bool
		expectedImports  []string
		unexpectedImport string
	}{
		{
			name: "add seeder imports to file with existing imports",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "goravel/database/seeders",
			expectError: false,
			expectedImports: []string{
				"goravel/database/seeders",
				"github.com/goravel/framework/contracts/database/seeder",
			},
		},
		{
			name: "add seeder imports when seeder import already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "goravel/database/seeders",
			expectError: false,
			expectedImports: []string{
				"goravel/database/seeders",
				"github.com/goravel/framework/contracts/database/seeder",
			},
		},
		{
			name: "add seeder imports when seeder package already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/database/seeders"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "goravel/database/seeders",
			expectError: false,
			expectedImports: []string{
				"goravel/database/seeders",
				"github.com/goravel/framework/contracts/database/seeder",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			err := addSeederImports(sourceFile, tt.pkg)

			if tt.expectError {
				s.Error(err)
				return
			}

			s.NoError(err)

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			for _, expectedImport := range tt.expectedImports {
				s.Contains(content, expectedImport, "Expected import %s to be present", expectedImport)
			}

			if tt.unexpectedImport != "" {
				s.NotContains(content, tt.unexpectedImport)
			}
		})
	}
}

func (s *ModifyActionsTestSuite) Test_foundationSetupSeeder() {
	tests := []struct {
		name           string
		initialContent string
		seederToAdd    string
		expectedResult string
	}{
		{
			name: "create WithSeeders when it doesn't exist",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			seederToAdd: "&seeders.DatabaseSeeder{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.DatabaseSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "append to existing WithSeeders",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.ExistingSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
			seederToAdd: "&seeders.DatabaseSeeder{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() {
	foundation.Setup().
		WithSeeders([]seeder.Seeder{
			&seeders.ExistingSeeder{},
			&seeders.DatabaseSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "skip non-foundation.Setup() statements",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
			seederToAdd: "&seeders.DatabaseSeeder{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "test.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			content, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			_, err = decorator.Parse(content)
			s.Require().NoError(err)

			// Apply the action
			err = GoFile(sourceFile).Find(match.FoundationSetup()).Modify(foundationSetupSeeder(tt.seederToAdd)).Apply()
			s.NoError(err)

			// Read the result
			resultContent, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			s.Equal(tt.expectedResult, resultContent)
		})
	}
}
