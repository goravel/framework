package paths

import (
	"path"
	"path/filepath"
	"runtime/debug"
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

func (s *PathsTestSuite) TestBootstrap() {
	support.Config.Paths.Bootstrap = "bootstrap"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Bootstrap()

	s.NotNil(result)
	s.Equal("bootstrap", result.Package())
	s.Equal("goravel/bootstrap", result.Import())
}

func (s *PathsTestSuite) TestConfig() {
	support.Config.Paths.Config = "config"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Config()

	s.NotNil(result)
	s.Equal("config", result.Package())
	s.Equal("goravel/config", result.Import())
}

func (s *PathsTestSuite) TestFacades() {
	support.Config.Paths.Facades = "app/facades"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Facades()

	s.NotNil(result)
	s.Equal("facades", result.Package())
	s.Equal("goravel/app/facades", result.Import())
}

func (s *PathsTestSuite) TestMain() {
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Main()

	s.NotNil(result)
	s.Equal("goravel", result.Package())
	// Main() passes mainPath as both path and main, so Import() returns "goravel/github.com/goravel/goravel"
	s.Equal("goravel/github.com/goravel/goravel", result.Import())
}

func (s *PathsTestSuite) TestMigrations() {
	support.Config.Paths.Migrations = "database/migrations"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Migrations()

	s.NotNil(result)
	s.Equal("migrations", result.Package())
	s.Equal("goravel/database/migrations", result.Import())
}

func (s *PathsTestSuite) TestModule() {
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Module()

	s.NotNil(result)
	// Module path depends on runtime/debug.ReadBuildInfo(), which may vary
	// Just verify it returns a non-nil Path
}

func (s *PathsTestSuite) TestRoutes() {
	support.Config.Paths.Routes = "routes"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Routes()

	s.NotNil(result)
	s.Equal("routes", result.Package())
	s.Equal("goravel/routes", result.Import())
}

func (s *PathsTestSuite) TestTests() {
	support.Config.Paths.Tests = "tests"
	mainPath := "github.com/goravel/goravel"

	paths := NewPaths(mainPath)
	result := paths.Tests()

	s.NotNil(result)
	s.Equal("tests", result.Package())
	s.Equal("goravel/tests", result.Import())
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

func (s *PathTestSuite) TestModulePath() {
	// Test Module() method behavior with runtime/debug
	mainPath := "github.com/goravel/goravel"
	paths := NewPaths(mainPath)

	result := paths.Module()
	s.NotNil(result)

	// Check if build info is available
	if info, ok := debug.ReadBuildInfo(); ok {
		if path.Ext(info.Path) == "/setup" {
			// If the path ends with "setup", Module should return parent directory
			expectedPath := path.Dir(info.Path)
			s.Equal(expectedPath, result.Import())
		}
	}
	// If build info is not available or doesn't end with setup, path should be empty
	// but the Path object should still be valid
	s.IsType(&Path{}, result)
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
