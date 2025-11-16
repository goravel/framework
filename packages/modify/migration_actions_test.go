package modify

import (
	"path/filepath"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
)

type MigrationActionsTestSuite struct {
	suite.Suite
}

func TestMigrationActionsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationActionsTestSuite))
}

func (s *MigrationActionsTestSuite) TestAddMigration() {
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
	"goravel/bootstrap"
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
	"goravel/bootstrap"
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
		s.Run(tt.name, func() {
			tempDir := s.T().TempDir()
			bootstrapDir := filepath.Join(tempDir, "bootstrap")

			appFile := filepath.Join(bootstrapDir, "app.go")
			migrationsFile := filepath.Join(bootstrapDir, "migrations.go")

			s.Require().NoError(supportfile.PutContent(appFile, tt.appContent))

			if tt.migrationsContent != "" {
				s.Require().NoError(supportfile.PutContent(migrationsFile, tt.migrationsContent))
			}

			// Override Config.Paths.App for testing
			originalAppPath := support.Config.Paths.App
			support.Config.Paths.App = appFile
			defer func() {
				support.Config.Paths.App = originalAppPath
			}()

			err := AddMigration(tt.pkg, tt.migration)

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

			// Verify migrations.go content if expected
			if tt.expectedMigrations != "" {
				migrationsContent, err := supportfile.GetContent(migrationsFile)
				s.Require().NoError(err)
				s.Equal(tt.expectedMigrations, migrationsContent)
			}
		})
	}
}

func (s *MigrationActionsTestSuite) Test_appendToExistingMigration() {
	tests := []struct {
		name              string
		initialContent    string
		migrationToAdd    string
		expectedArgsCount int
	}{
		{
			name: "append to existing WithMigrations call",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/foundation"
	"goravel/database/migrations"
)

func Boot() {
	foundation.Setup().
		WithMigrations([]schema.Migration{
			&migrations.ExistingMigration{},
		}).Run()
}`,
			migrationToAdd:    "&migrations.CreateUsersTable{}",
			expectedArgsCount: 2,
		},
		{
			name: "append to empty migration array",
			initialContent: `package test

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/foundation"
	"goravel/database/migrations"
)

func Boot() {
	foundation.Setup().
		WithMigrations([]schema.Migration{}).Run()
}`,
			migrationToAdd:    "&migrations.CreateUsersTable{}",
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

			// Find the WithMigrations call
			var withMigrationsCall *dst.CallExpr
			dst.Inspect(file, func(n dst.Node) bool {
				if call, ok := n.(*dst.CallExpr); ok {
					if sel, ok := call.Fun.(*dst.SelectorExpr); ok {
						if sel.Sel.Name == "WithMigrations" {
							withMigrationsCall = call
							return false
						}
					}
				}
				return true
			})

			s.Require().NotNil(withMigrationsCall, "WithMigrations call not found")

			migrationExpr := MustParseExpr(tt.migrationToAdd).(dst.Expr)
			appendToExistingMigration(withMigrationsCall, migrationExpr)

			// Verify the migration was appended
			s.Require().Len(withMigrationsCall.Args, 1)
			compositeLit, ok := withMigrationsCall.Args[0].(*dst.CompositeLit)
			s.Require().True(ok)
			s.Equal(tt.expectedArgsCount, len(compositeLit.Elts))
		})
	}
}

func (s *MigrationActionsTestSuite) Test_addMigrationImports() {
	tests := []struct {
		name             string
		initialContent   string
		pkg              string
		expectError      bool
		expectedImports  []string
		unexpectedImport string
	}{
		{
			name: "add migration imports to file with existing imports",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "goravel/database/migrations",
			expectError: false,
			expectedImports: []string{
				"goravel/database/migrations",
				"github.com/goravel/framework/contracts/database/schema",
			},
		},
		{
			name: "add migration imports when schema import already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "goravel/database/migrations",
			expectError: false,
			expectedImports: []string{
				"goravel/database/migrations",
				"github.com/goravel/framework/contracts/database/schema",
			},
		},
		{
			name: "add migration imports when migration package already exists",
			initialContent: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/database/migrations"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			pkg:         "goravel/database/migrations",
			expectError: false,
			expectedImports: []string{
				"goravel/database/migrations",
				"github.com/goravel/framework/contracts/database/schema",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			sourceFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(sourceFile, tt.initialContent))

			err := addMigrationImports(sourceFile, tt.pkg)

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

func (s *MigrationActionsTestSuite) Test_foundationSetupMigration() {
	tests := []struct {
		name           string
		initialContent string
		migrationToAdd string
		expectedResult string
	}{
		{
			name: "create WithMigrations when it doesn't exist",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			migrationToAdd: "&migrations.CreateUsersTable{}",
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMigrations([]schema.Migration{
			&migrations.CreateUsersTable{},
		}).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "append to existing WithMigrations",
			initialContent: `package test

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
			migrationToAdd: "&migrations.CreateUsersTable{}",
			expectedResult: `package test

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
			migrationToAdd: "&migrations.CreateUsersTable{}",
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
			err = GoFile(sourceFile).Find(match.FoundationSetup()).Modify(foundationSetupMigrationInline(tt.migrationToAdd)).Apply()
			s.NoError(err)

			// Read the result
			resultContent, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			s.Equal(tt.expectedResult, resultContent)
		})
	}
}

func (s *MigrationActionsTestSuite) Test_checkWithMigrationsExists() {
	tests := []struct {
		name     string
		content  string
		expected bool
		wantErr  bool
	}{
		{
			name: "WithMigrations exists in chain",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/bootstrap"
)

func Boot() {
	foundation.Setup().WithMigrations(Migrations()).Run()
}
`,
			expected: true,
		},
		{
			name: "WithMigrations exists with inline array",
			content: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().WithMigrations([]schema.Migration{}).Run()
}
`,
			expected: true,
		},
		{
			name: "WithMigrations doesn't exist",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			expected: false,
		},
		{
			name: "WithMigrations doesn't exist in complex chain",
			content: `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Run()
}
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "app.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.content))

			result, err := checkWithMigrationsExists(tempFile)

			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *MigrationActionsTestSuite) Test_createMigrationsFile() {
	tests := []struct {
		name            string
		expectedContent string
	}{
		{
			name: "create migrations.go file with correct structure",
			expectedContent: `package bootstrap

import "github.com/goravel/framework/contracts/database/schema"

func Migrations() []schema.Migration {
	return []schema.Migration{}
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "migrations.go")

			err := createMigrationsFile(tempFile)
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedContent, content)
		})
	}
}

func (s *MigrationActionsTestSuite) Test_addMigrationToMigrationsFile() {
	tests := []struct {
		name            string
		initialContent  string
		pkg             string
		migration       string
		expectedContent string
	}{
		{
			name: "add migration to empty Migrations() function",
			initialContent: `package bootstrap

import "github.com/goravel/framework/contracts/database/schema"

func Migrations() []schema.Migration {
	return []schema.Migration{}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedContent: `package bootstrap

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
			name: "add migration to existing Migrations() function",
			initialContent: `package bootstrap

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
			migration: "&migrations.CreatePostsTable{}",
			expectedContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
		&migrations.CreatePostsTable{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tempFile := filepath.Join(s.T().TempDir(), "migrations.go")
			s.Require().NoError(supportfile.PutContent(tempFile, tt.initialContent))

			err := addMigrationToMigrationsFile(tempFile, tt.pkg, tt.migration)
			s.NoError(err)

			content, err := supportfile.GetContent(tempFile)
			s.Require().NoError(err)
			s.Equal(tt.expectedContent, content)
		})
	}
}

func (s *MigrationActionsTestSuite) Test_foundationSetupMigrationWithFunction() {
	tests := []struct {
		name           string
		initialContent string
		expectedResult string
	}{
		{
			name: "add WithMigrations(Migrations()) when it doesn't exist",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`,
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithMigrations(Migrations()).WithConfig(config.Boot).Run()
}
`,
		},
		{
			name: "add WithMigrations(Migrations()) at the beginning of chain",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`,
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithMigrations(Migrations()).Run()
}
`,
		},
		{
			name: "skip non-foundation.Setup() statements",
			initialContent: `package test

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	app := foundation.NewApplication()
	app.Run()
}
`,
			expectedResult: `package test

import (
	"github.com/goravel/framework/foundation"
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
			err = GoFile(sourceFile).Find(match.FoundationSetup()).Modify(foundationSetupMigrationWithFunction()).Apply()
			s.NoError(err)

			// Read the result
			resultContent, err := supportfile.GetContent(sourceFile)
			s.Require().NoError(err)

			s.Equal(tt.expectedResult, resultContent)
		})
	}
}
