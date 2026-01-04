package paths

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support"
)

type PathsTestSuite struct {
	suite.Suite
	originalPaths support.Paths
}

func TestPathsTestSuite(t *testing.T) {
	suite.Run(t, new(PathsTestSuite))
}

func (s *PathsTestSuite) SetupTest() {
	s.originalPaths = support.Config.Paths
}

func (s *PathsTestSuite) TearDownTest() {
	support.Config.Paths = s.originalPaths
}

func (s *PathsTestSuite) TestNewPaths() {
	tests := []struct {
		name     string
		mainPath string
	}{
		{
			name:     "with simple main path",
			mainPath: "github.com/goravel/goravel",
		},
		{
			name:     "with empty main path",
			mainPath: "",
		},
		{
			name:     "with complex main path",
			mainPath: "github.com/organization/project/submodule",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			paths := NewPaths(tt.mainPath)
			s.NotNil(paths)
			s.Equal(tt.mainPath, paths.mainPath)
		})
	}
}

func (s *PathsTestSuite) TestApp() {
	support.Config.Paths.App = "app"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.App()

	s.NotNil(result)
	s.Equal("app", result.Package())
	s.Equal("goravel/app", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("app", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "app")
}

func (s *PathsTestSuite) TestBootstrap() {
	support.Config.Paths.Bootstrap = "bootstrap"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Bootstrap()

	s.NotNil(result)
	s.Equal("bootstrap", result.Package())
	s.Equal("goravel/bootstrap", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("bootstrap", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "bootstrap")
}

func (s *PathsTestSuite) TestConfig() {
	support.Config.Paths.Config = "config"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Config()

	s.NotNil(result)
	s.Equal("config", result.Package())
	s.Equal("goravel/config", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("config", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "config")
}

func (s *PathsTestSuite) TestDatabase() {
	support.Config.Paths.Database = "database"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Database()

	s.NotNil(result)
	s.Equal("database", result.Package())
	s.Equal("goravel/database", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("database", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "database")
}

func (s *PathsTestSuite) TestFacades() {
	support.Config.Paths.Facades = "app/facades"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Facades()

	s.NotNil(result)
	s.Equal("facades", result.Package())
	s.Equal("goravel/app/facades", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal(filepath.Join("app", "facades"), result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), filepath.Join("app", "facades"))
}

func (s *PathsTestSuite) TestLang() {
	support.Config.Paths.Lang = "lang"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Lang()

	s.NotNil(result)
	s.Equal("lang", result.Package())
	s.Equal("goravel/lang", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("lang", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "lang")
}

func (s *PathsTestSuite) TestMain() {
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Main()

	s.NotNil(result)
	s.Equal("goravel", result.Package())
	s.Equal("goravel", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("", result.String())
	s.True(filepath.IsAbs(result.Abs()))
}

func (s *PathsTestSuite) TestMigrations() {
	support.Config.Paths.Migrations = "database/migrations"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Migrations()

	s.NotNil(result)
	s.Equal("migrations", result.Package())
	s.Equal("goravel/database/migrations", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal(filepath.Join("database", "migrations"), result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), filepath.Join("database", "migrations"))
}

func (s *PathsTestSuite) TestModels() {
	support.Config.Paths.Models = "app/models"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Models()

	s.NotNil(result)
	s.Equal("models", result.Package())
	s.Equal("goravel/app/models", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal(filepath.Join("app", "models"), result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), filepath.Join("app", "models"))
}

func (s *PathsTestSuite) TestModule() {
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Module()

	s.NotNil(result)
	// Module path depends on runtime/debug.ReadBuildInfo(), which may vary
	// Just verify it returns a non-nil Path
}

func (s *PathsTestSuite) TestPublic() {
	support.Config.Paths.Public = "public"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Public()

	s.NotNil(result)
	s.Equal("public", result.Package())
	s.Equal("goravel/public", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("public", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "public")
}

func (s *PathsTestSuite) TestResources() {
	support.Config.Paths.Resources = "resources"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Resources()

	s.NotNil(result)
	s.Equal("resources", result.Package())
	s.Equal("goravel/resources", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("resources", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "resources")
}

func (s *PathsTestSuite) TestRoutes() {
	support.Config.Paths.Routes = "routes"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Routes()

	s.NotNil(result)
	s.Equal("routes", result.Package())
	s.Equal("goravel/routes", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("routes", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "routes")
}

func (s *PathsTestSuite) TestStorage() {
	support.Config.Paths.Storage = "storage"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Storage()

	s.NotNil(result)
	s.Equal("storage", result.Package())
	s.Equal("goravel/storage", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("storage", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "storage")
}

func (s *PathsTestSuite) TestTests() {
	support.Config.Paths.Tests = "tests"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Tests()

	s.NotNil(result)
	s.Equal("tests", result.Package())
	s.Equal("goravel/tests", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal("tests", result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), "tests")
}

func (s *PathsTestSuite) TestViews() {
	support.Config.Paths.Views = "resources/views"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Views()

	s.NotNil(result)
	s.Equal("views", result.Package())
	s.Equal("goravel/resources/views", result.Import())
	s.False(filepath.IsAbs(result.String()))
	s.Equal(filepath.Join("resources", "views"), result.String())
	s.True(filepath.IsAbs(result.Abs()))
	s.Contains(result.Abs(), filepath.Join("resources", "views"))
}

type PathTestSuite struct {
	suite.Suite
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, new(PathTestSuite))
}

func (s *PathTestSuite) TestNewPath() {
	tests := []struct {
		name     string
		path     string
		main     string
		isModule bool
	}{
		{
			name:     "with simple path",
			path:     "app/http/controllers",
			main:     "github.com/goravel/goravel",
			isModule: false,
		},
		{
			name:     "with empty path",
			path:     "",
			main:     "github.com/goravel/goravel",
			isModule: false,
		},
		{
			name:     "with module path",
			path:     "github.com/goravel/framework/auth",
			main:     "github.com/goravel/goravel",
			isModule: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := NewPath(tt.path, tt.main, tt.isModule)
			s.NotNil(result)
			s.Equal(tt.path, result.path)
			s.Equal(tt.main, result.main)
			s.Equal(tt.isModule, result.isModule)
		})
	}
}

func (s *PathTestSuite) TestPackage() {
	tests := []struct {
		name     string
		path     string
		main     string
		expected string
	}{
		{
			name:     "with sub-package path",
			path:     "app/http/controllers",
			main:     "github.com/goravel/goravel",
			expected: "controllers",
		},
		{
			name:     "with single segment path",
			path:     "config",
			main:     "github.com/goravel/goravel",
			expected: "config",
		},
		{
			name:     "with empty path returns main package",
			path:     "",
			main:     "github.com/goravel/goravel",
			expected: "goravel",
		},
		{
			name:     "with nested path",
			path:     "app/http/controllers/admin",
			main:     "github.com/user/project",
			expected: "admin",
		},
		{
			name:     "with windows-style path",
			path:     "app\\http\\controllers",
			main:     "github.com/goravel/goravel",
			expected: "controllers",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := NewPath(tt.path, tt.main, false)
			result := p.Package()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *PathTestSuite) TestImport() {
	tests := []struct {
		name     string
		path     string
		main     string
		isModule bool
		expected string
	}{
		{
			name:     "with sub-package path",
			path:     "app/http/controllers",
			main:     "github.com/goravel/goravel",
			isModule: false,
			expected: "goravel/app/http/controllers",
		},
		{
			name:     "with empty path returns main import",
			path:     "",
			main:     "github.com/goravel/goravel",
			isModule: false,
			expected: "goravel",
		},
		{
			name:     "with single segment path",
			path:     "config",
			main:     "github.com/user/project",
			isModule: false,
			expected: "project/config",
		},
		{
			name:     "with module path",
			path:     "github.com/goravel/framework/auth",
			main:     "github.com/goravel/goravel",
			isModule: true,
			expected: "github.com/goravel/framework/auth",
		},
		{
			name:     "with nested sub-package",
			path:     "app/http/controllers/admin",
			main:     "github.com/organization/app",
			isModule: false,
			expected: "app/app/http/controllers/admin",
		},
		{
			name:     "with windows-style path",
			path:     "app\\http\\controllers",
			main:     "github.com/goravel/goravel",
			isModule: false,
			expected: "goravel/app/http/controllers",
		},
		{
			name:     "with complex main path",
			path:     "routes",
			main:     "github.com/company/division/product",
			isModule: false,
			expected: "product/routes",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := NewPath(tt.path, tt.main, tt.isModule)
			result := p.Import()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *PathTestSuite) TestString() {
	tests := []struct {
		name            string
		path            string
		main            string
		additionalPaths []string
		expected        string
	}{
		{
			name:            "without additional paths",
			path:            "app/http/controllers",
			main:            "github.com/goravel/goravel",
			additionalPaths: nil,
			expected:        filepath.Join("app", "http", "controllers"),
		},
		{
			name:            "with single additional path",
			path:            "app/http",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"controllers"},
			expected:        filepath.Join("app", "http", "controllers"),
		},
		{
			name:            "with multiple additional paths",
			path:            "app",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"http", "controllers", "user.go"},
			expected:        filepath.Join("app", "http", "controllers", "user.go"),
		},
		{
			name:            "with empty path and additional paths",
			path:            "",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"config", "app.go"},
			expected:        filepath.Join("config", "app.go"),
		},
		{
			name:            "with windows-style path",
			path:            "app\\models",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"user.go"},
			expected:        filepath.Join("app", "models", "user.go"),
		},
		{
			name:            "with empty path",
			path:            "",
			main:            "github.com/goravel/goravel",
			additionalPaths: nil,
			expected:        "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := NewPath(tt.path, tt.main, false)
			result := p.String(tt.additionalPaths...)

			s.False(filepath.IsAbs(result), "Result should be a relative path")
			s.Equal(tt.expected, result)
		})
	}
}

func (s *PathTestSuite) TestAbs() {
	originalRelativePath := support.RelativePath
	defer func() {
		support.RelativePath = originalRelativePath
	}()

	support.RelativePath = "."

	tests := []struct {
		name            string
		path            string
		main            string
		additionalPaths []string
		expectContain   string
	}{
		{
			name:            "without additional paths",
			path:            "app/http/controllers",
			main:            "github.com/goravel/goravel",
			additionalPaths: nil,
			expectContain:   filepath.Join("app", "http", "controllers"),
		},
		{
			name:            "with single additional path",
			path:            "app/http",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"controllers"},
			expectContain:   filepath.Join("app", "http", "controllers"),
		},
		{
			name:            "with multiple additional paths",
			path:            "app",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"http", "controllers", "user.go"},
			expectContain:   filepath.Join("app", "http", "controllers", "user.go"),
		},
		{
			name:            "with empty path and additional paths",
			path:            "",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"config", "app.go"},
			expectContain:   filepath.Join("config", "app.go"),
		},
		{
			name:            "with windows-style path",
			path:            "app\\models",
			main:            "github.com/goravel/goravel",
			additionalPaths: []string{"user.go"},
			expectContain:   filepath.Join("app", "models", "user.go"),
		},
		{
			name:            "with empty path",
			path:            "",
			main:            "github.com/goravel/goravel",
			additionalPaths: nil,
			expectContain:   "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := NewPath(tt.path, tt.main, false)
			result := p.Abs(tt.additionalPaths...)

			s.True(filepath.IsAbs(result), "Result should be an absolute path")
			if tt.expectContain != "" {
				s.Contains(result, tt.expectContain, "Result should contain expected path")
			}
		})
	}
}

func (s *PathTestSuite) TestPkg() {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "with multi-segment path",
			path:     "app/http/controllers",
			expected: "controllers",
		},
		{
			name:     "with single segment",
			path:     "config",
			expected: "config",
		},
		{
			name:     "with empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "with trailing slash",
			path:     "app/http/",
			expected: "http",
		},
		{
			name:     "with windows-style path",
			path:     "app\\http\\controllers",
			expected: "controllers",
		},
		{
			name:     "with deeply nested path",
			path:     "app/http/controllers/admin/users",
			expected: "users",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := pkg(tt.path)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *PathTestSuite) TestModuleImportBehavior() {
	tests := []struct {
		name        string
		modulePath  string
		mainPath    string
		expectedPkg string
	}{
		{
			name:        "module with framework path",
			modulePath:  "github.com/goravel/framework/auth",
			mainPath:    "github.com/goravel/goravel",
			expectedPkg: "auth",
		},
		{
			name:        "module with nested path",
			modulePath:  "github.com/goravel/framework/database/orm",
			mainPath:    "github.com/goravel/goravel",
			expectedPkg: "orm",
		},
		{
			name:        "empty module path returns main package",
			modulePath:  "",
			mainPath:    "github.com/goravel/goravel",
			expectedPkg: "goravel",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := NewPath(tt.modulePath, tt.mainPath, true)
			s.Equal(tt.expectedPkg, p.Package())

			if tt.modulePath != "" {
				// Module paths should return the path directly
				s.Equal(tt.modulePath, p.Import())
			} else {
				// Empty module path should return main import
				s.Equal("goravel", p.Import())
			}
		})
	}
}

func (s *PathTestSuite) TestPathWithCustomConfigs() {
	originalPaths := support.Config.Paths
	defer func() {
		support.Config.Paths = originalPaths
	}()

	// Test with custom paths
	support.Config.Paths.Bootstrap = "custom/bootstrap"
	support.Config.Paths.Config = "custom/config"
	support.Config.Paths.Facades = "custom/app/facades"
	support.Config.Paths.Migrations = "custom/db/migrations"
	support.Config.Paths.Routes = "custom/routes"
	support.Config.Paths.Tests = "custom/tests"

	mainPath := "github.com/goravel/goravel"
	paths := NewPaths(mainPath)

	s.Equal("bootstrap", paths.Bootstrap().Package())
	s.Equal("goravel/custom/bootstrap", paths.Bootstrap().Import())

	s.Equal("config", paths.Config().Package())
	s.Equal("goravel/custom/config", paths.Config().Import())

	s.Equal("facades", paths.Facades().Package())
	s.Equal("goravel/custom/app/facades", paths.Facades().Import())

	s.Equal("migrations", paths.Migrations().Package())
	s.Equal("goravel/custom/db/migrations", paths.Migrations().Import())

	s.Equal("routes", paths.Routes().Package())
	s.Equal("goravel/custom/routes", paths.Routes().Import())

	s.Equal("tests", paths.Tests().Package())
	s.Equal("goravel/custom/tests", paths.Tests().Import())
}

func (s *PathTestSuite) TestPathWithDifferentMainPaths() {
	tests := []struct {
		name         string
		mainPath     string
		subPath      string
		expectedPkg  string
		expectedImpt string
	}{
		{
			name:         "github.com path",
			mainPath:     "github.com/goravel/goravel",
			subPath:      "app/http",
			expectedPkg:  "http",
			expectedImpt: "goravel/app/http",
		},
		{
			name:         "gitlab.com path",
			mainPath:     "gitlab.com/user/project",
			subPath:      "controllers",
			expectedPkg:  "controllers",
			expectedImpt: "project/controllers",
		},
		{
			name:         "single segment main",
			mainPath:     "myapp",
			subPath:      "routes",
			expectedPkg:  "routes",
			expectedImpt: "myapp/routes",
		},
		{
			name:         "deeply nested main",
			mainPath:     "git.company.com/team/division/project",
			subPath:      "models",
			expectedPkg:  "models",
			expectedImpt: "project/models",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p := NewPath(tt.subPath, tt.mainPath, false)
			s.Equal(tt.expectedPkg, p.Package())
			s.Equal(tt.expectedImpt, p.Import())
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		name          string
		relativePath  string
		paths         []string
		expectContain string
	}{
		{
			name:          "single path",
			relativePath:  ".",
			paths:         []string{"test.txt"},
			expectContain: "test.txt",
		},
		{
			name:          "multiple paths",
			relativePath:  ".",
			paths:         []string{"app", "controllers", "user.go"},
			expectContain: filepath.Join("app", "controllers", "user.go"),
		},
		{
			name:          "empty paths",
			relativePath:  ".",
			paths:         []string{},
			expectContain: "",
		},
		{
			name:          "nested paths",
			relativePath:  ".",
			paths:         []string{"foo", "bar", "baz", "file.txt"},
			expectContain: filepath.Join("foo", "bar", "baz", "file.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			support.RelativePath = tt.relativePath
			result := Abs(tt.paths...)

			// The result should be an absolute path
			assert.True(t, filepath.IsAbs(result))

			// The result should contain the expected path components
			if tt.expectContain != "" {
				assert.Contains(t, result, tt.expectContain)
			}
		})
	}
}

func TestToSlice(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "Simple path with forward slashes",
			path:     "app/http/controllers",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Windows path with backslashes",
			path:     "app\\http\\controllers",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with leading and trailing slashes",
			path:     "/app/http/controllers/",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Mixed slashes with leading and trailing",
			path:     "\\app\\http\\controllers\\",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Single directory",
			path:     "app",
			expected: []string{"app"},
		},
		{
			name:     "Deep nested path",
			path:     "app/http/controllers/api/v1/users",
			expected: []string{"app", "http", "controllers", "api", "v1", "users"},
		},
		{
			name:     "Empty string",
			path:     "",
			expected: nil,
		},
		{
			name:     "Root forward slash",
			path:     "/",
			expected: nil,
		},
		{
			name:     "Root backslash",
			path:     "\\",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSlice(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
