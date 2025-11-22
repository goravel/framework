package modify

import (
	"bytes"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
)

func TestAddCommand(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		commandsContent   string // empty if file doesn't exist
		pkg               string
		command           string
		expectedApp       string
		expectedCommands  string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add command when WithCommands doesn't exist and commands.go doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.ExampleCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExampleCommand{},
	}
}
`,
		},
		{
			name: "add command when WithCommands exists with Commands() and commands.go exists",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
			commandsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).WithConfig(config.Boot).Run()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
		&commands.NewCommand{},
	}
}
`,
		},
		{
			name: "add command when WithCommands exists with inline array",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.ExistingCommand{},
		}).WithConfig(config.Boot).Run()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.ExistingCommand{},
			&commands.NewCommand{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "error when commands.go exists but WithCommands doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			commandsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			wantErr: true,
		},
		{
			name: "add command when WithCommands doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.FirstCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithCommands(Commands()).Run()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.FirstCommand{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			commandsFile := filepath.Join(bootstrapDir, "commands.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.commandsContent != "" {
				assert.NoError(t, supportfile.PutContent(commandsFile, tt.commandsContent))
			}

			err := AddCommand(tt.pkg, tt.command)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify commands.go content if expected
			if tt.expectedCommands != "" {
				commandsContent, err := supportfile.GetContent(commandsFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCommands, commandsContent)
			}
		})
	}
}

func TestAddMiddleware(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			sourceFile := filepath.Join(bootstrapDir, "app.go")

			assert.NoError(t, supportfile.PutContent(sourceFile, tt.content))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			err := AddMiddleware(tt.pkg, tt.mw)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			content, err := supportfile.GetContent(sourceFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, content)
		})
	}
}

func TestAddMigration(t *testing.T) {
	tests := []struct {
		name               string
		appContent         string
		migrationsContent  string // empty if file doesn't exist
		pkg                string
		migration          string
		expectedApp        string
		expectedMigrations string // empty if file shouldn't be created
		wantErr            bool
		expectedErrString  string
	}{
		{
			name: "add migration when WithMigrations doesn't exist and migrations.go doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMigrations(Migrations()).WithConfig(config.Boot).Run()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
	}
}
`,
		},
		{
			name: "add migration when WithMigrations exists with Migrations() and migrations.go exists",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMigrations(Migrations()).WithConfig(config.Boot).Run()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMigrations(Migrations()).WithConfig(config.Boot).Run()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
		&migrations.CreateUsersTable{},
	}
}
`,
		},
		{
			name: "add migration when WithMigrations exists with inline array",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/migrations"
)

func Boot() {
	foundation.Setup().
		WithMigrations([]schema.Migration{
			&migrations.ExistingMigration{},
		}).WithConfig(config.Boot).Run()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/migrations"
)

func Boot() {
	foundation.Setup().
		WithMigrations([]schema.Migration{
			&migrations.ExistingMigration{},
			&migrations.CreateUsersTable{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "error when migrations.go exists but WithMigrations doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			wantErr:   true,
		},
		{
			name: "add migration when WithMigrations doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreatePostsTable{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithMigrations(Migrations()).Run()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreatePostsTable{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			migrationsFile := filepath.Join(bootstrapDir, "migrations.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.migrationsContent != "" {
				assert.NoError(t, supportfile.PutContent(migrationsFile, tt.migrationsContent))
			}

			err := AddMigration(tt.pkg, tt.migration)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify migrations.go content if expected
			if tt.expectedMigrations != "" {
				migrationsContent, err := supportfile.GetContent(migrationsFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMigrations, migrationsContent)
			}
		})
	}
}

func TestAddProvider(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			providersFile := filepath.Join(bootstrapDir, "providers.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.providersContent != "" {
				require.NoError(t, supportfile.PutContent(providersFile, tt.providersContent))
			}

			err := AddProvider(tt.pkg, tt.provider)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrString != "" {
					require.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			require.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify providers.go content if expected
			if tt.expectedProviders != "" {
				providersContent, err := supportfile.GetContent(providersFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedProviders, providersContent)
			}
		})
	}
}

func TestRemoveProvider(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		providersContent  string // empty if file doesn't exist
		pkg               string
		provider          string
		expectedApp       string
		expectedProviders string // expected content after removal, empty if file doesn't exist
	}{
		{
			name: "remove provider from providers.go when multiple providers exist",
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
		&providers.RouteServiceProvider{},
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
		&providers.RouteServiceProvider{},
	}
}
`,
		},
		{
			name: "remove provider from inline array when multiple providers exist",
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
			&providers.AppServiceProvider{},
			&providers.RouteServiceProvider{},
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
			&providers.RouteServiceProvider{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "remove last provider from providers.go",
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
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{}
}
`,
		},
		{
			name: "remove provider from different package",
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
	"github.com/goravel/redis"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
		&redis.ServiceProvider{},
	}
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
			name: "no-op when WithProviders doesn't exist",
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
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "no-op when provider doesn't exist in the list",
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
		&providers.RouteServiceProvider{},
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
		&providers.RouteServiceProvider{},
	}
}
`,
		},
		{
			name: "remove provider but keep import when another provider from same package exists",
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
		&providers.RouteServiceProvider{},
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
		&providers.RouteServiceProvider{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			providersFile := filepath.Join(bootstrapDir, "providers.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.providersContent != "" {
				require.NoError(t, supportfile.PutContent(providersFile, tt.providersContent))
			}

			err := RemoveProvider(tt.pkg, tt.provider)
			require.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify providers.go content if expected
			if tt.expectedProviders != "" {
				providersContent, err := supportfile.GetContent(providersFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedProviders, providersContent)
			}
		})
	}
}

func TestAddSeeder(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		seedersContent    string // empty if file doesn't exist
		pkg               string
		seeder            string
		expectedApp       string
		expectedSeeders   string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add seeder when WithSeeders doesn't exist and seeders.go doesn't exist",
			appContent: `package bootstrap

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
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithSeeders(Seeders()).WithConfig(config.Boot).Run()
}
`,
			expectedSeeders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
	}
}
`,
		},
		{
			name: "add seeder when WithSeeders exists with Seeders() and seeders.go exists",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithSeeders(Seeders()).WithConfig(config.Boot).Run()
}
`,
			seedersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.ExistingSeeder{},
	}
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.NewSeeder{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithSeeders(Seeders()).WithConfig(config.Boot).Run()
}
`,
			expectedSeeders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.ExistingSeeder{},
		&seeders.NewSeeder{},
	}
}
`,
		},
		{
			name: "add seeder when WithSeeders exists with inline array",
			appContent: `package bootstrap

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
			seeder: "&seeders.NewSeeder{}",
			expectedApp: `package bootstrap

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
			&seeders.NewSeeder{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "error when seeders.go exists but WithSeeders doesn't exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			seedersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.ExistingSeeder{},
	}
}
`,
			pkg:               "goravel/database/seeders",
			seeder:            "&seeders.NewSeeder{}",
			wantErr:           true,
			expectedErrString: "seeders.go already exists but WithSeeders is not registered in foundation.Setup()",
		},
		{
			name: "add seeder when WithSeeders doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.FirstSeeder{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithSeeders(Seeders()).Run()
}
`,
			expectedSeeders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.FirstSeeder{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			seedersFile := filepath.Join(bootstrapDir, "seeders.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.seedersContent != "" {
				assert.NoError(t, supportfile.PutContent(seedersFile, tt.seedersContent))
			}

			err := AddSeeder(tt.pkg, tt.seeder)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify seeders.go content if expected
			if tt.expectedSeeders != "" {
				seedersContent, err := supportfile.GetContent(seedersFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSeeders, seedersContent)
			}
		})
	}
}

func TestExprExists(t *testing.T) {
	assert.NotPanics(t, func() {
		t.Run("expr exists", func(t *testing.T) {
			assert.True(t,
				ExprExists(
					MustParseExpr("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
			assert.NotEqual(t, -1,
				ExprIndex(
					MustParseExpr("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
		})
		t.Run("expr does not exist", func(t *testing.T) {
			assert.False(t,
				ExprExists(
					MustParseExpr("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
			assert.Equal(t, -1,
				ExprIndex(
					MustParseExpr("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
		})
	})
}

func TestUsesImport(t *testing.T) {
	df, err := decorator.Parse(`package main
import (
    "fmt"        
    mylog "log"
)

func main() {
    fmt.Println("hello")
}`)
	require.NoError(t, err)
	require.NotNil(t, df)

	assert.True(t, IsUsingImport(df, "fmt"))
	assert.False(t, IsUsingImport(df, "log", "mylog"))
}

func TestKeyExists(t *testing.T) {
	assert.NotPanics(t, func() {
		t.Run("key exists", func(t *testing.T) {
			assert.True(t,
				KeyExists(
					MustParseExpr(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
			assert.NotEqual(t, -1,
				KeyIndex(
					MustParseExpr(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
		})
		t.Run("key does not exist", func(t *testing.T) {
			assert.False(t,
				KeyExists(
					MustParseExpr(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
			assert.Equal(t, -1,
				KeyIndex(
					MustParseExpr(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
		})
	})
}

func TestMustParseStatement(t *testing.T) {
	t.Run("parse failed", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParseExpr("var invalid:=syntax")
		})
	})

	t.Run("parse success", func(t *testing.T) {
		assert.NotPanics(t, func() {
			assert.NotNil(t, MustParseExpr(`struct{x *int}`))
		})
	})
}

func TestWrapNewline(t *testing.T) {
	src := `package main

var value = 1
var _ = map[string]any{"key": &value, "func": func() bool { return true }}
`

	df, err := decorator.Parse(src)
	assert.NoError(t, err)

	// without WrapNewline
	var buf bytes.Buffer
	assert.NoError(t, decorator.Fprint(&buf, df))
	assert.Equal(t, src, buf.String())

	// with WrapNewline
	WrapNewline(df)
	buf.Reset()
	assert.NoError(t, decorator.Fprint(&buf, df))
	assert.NotEqual(t, src, buf.String())
	assert.Equal(t, `package main

var value = 1
var _ = map[string]any{
	"key": &value,
	"func": func() bool {
		return true
	},
}
`, buf.String())

}
