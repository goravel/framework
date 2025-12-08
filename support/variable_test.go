package support

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathToSlice(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "Unix path with forward slashes",
			path:     "app/http/controllers",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Windows path with backslashes",
			path:     "app\\http\\controllers",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with leading forward slash",
			path:     "/app/http/controllers",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with trailing forward slash",
			path:     "app/http/controllers/",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with leading and trailing forward slashes",
			path:     "/app/http/controllers/",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with leading backslash",
			path:     "\\app\\http\\controllers",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with trailing backslash",
			path:     "app\\http\\controllers\\",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Path with leading and trailing backslashes",
			path:     "\\app\\http\\controllers\\",
			expected: []string{"app", "http", "controllers"},
		},
		{
			name:     "Single directory",
			path:     "app",
			expected: []string{"app"},
		},
		{
			name:     "Single directory with leading slash",
			path:     "/app",
			expected: []string{"app"},
		},
		{
			name:     "Single directory with trailing slash",
			path:     "app/",
			expected: []string{"app"},
		},
		{
			name:     "Empty string",
			path:     "",
			expected: []string{""},
		},
		{
			name:     "Root forward slash",
			path:     "/",
			expected: []string{""},
		},
		{
			name:     "Root backslash",
			path:     "\\",
			expected: []string{""},
		},
		{
			name:     "Deep nested path with forward slashes",
			path:     "app/http/controllers/api/v1/users",
			expected: []string{"app", "http", "controllers", "api", "v1", "users"},
		},
		{
			name:     "Deep nested path with backslashes",
			path:     "app\\http\\controllers\\api\\v1\\users",
			expected: []string{"app", "http", "controllers", "api", "v1", "users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathToSlice(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
