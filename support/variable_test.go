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
			result := PathToSlice(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
