package internals

import (
	"path/filepath"
	"testing"

	"github.com/goravel/framework/support"
	"github.com/stretchr/testify/assert"
)

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

func TestBootstrapApp(t *testing.T) {
	tests := []struct {
		name          string
		relativePath  string
		bootstrapPath string
		expectContain string
	}{
		{
			name:          "default bootstrap path",
			relativePath:  ".",
			bootstrapPath: "bootstrap",
			expectContain: filepath.Join("bootstrap", "app.go"),
		},
		{
			name:          "custom bootstrap path",
			relativePath:  ".",
			bootstrapPath: "custom/bootstrap",
			expectContain: filepath.Join("custom", "bootstrap", "app.go"),
		},
		{
			name:          "bootstrap with single directory",
			relativePath:  ".",
			bootstrapPath: "boot",
			expectContain: filepath.Join("boot", "app.go"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			support.RelativePath = tt.relativePath
			support.Config.Paths.Bootstrap = tt.bootstrapPath

			result := BootstrapApp()

			// The result should be an absolute path
			assert.True(t, filepath.IsAbs(result))

			// The result should contain the expected path components
			assert.Contains(t, result, tt.expectContain)

			// The result should end with app.go
			assert.Contains(t, result, "app.go")
		})
	}
}

func TestFacadesPath(t *testing.T) {
	support.RelativePath = "." // Set to current dir for test
	result := Facades("foo.txt")
	expected := Abs(".", "app", "facades", "foo.txt")

	assert.Equal(t, expected, result)
}

func TestPath(t *testing.T) {
	support.RelativePath = "."
	result := Path("foo", "bar.txt")
	expected := Abs(".", "app", "foo", "bar.txt")

	assert.Equal(t, expected, result)
}

func TestPathToSlice(t *testing.T) {
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
			result := ToSlice(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
