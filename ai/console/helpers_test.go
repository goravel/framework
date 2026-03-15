package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectGoravelVersionFrom(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		hasError bool
	}{
		{
			name:     "valid go.mod",
			content:  "module example\n\nrequire github.com/goravel/framework v1.17.3\n",
			expected: "v1.17",
		},
		{
			name:     "malformed version string",
			content:  "module example\n\nrequire github.com/goravel/framework vX.Y.Z\n",
			hasError: true,
		},
		{
			name:     "framework not found",
			content:  "module example\n\nrequire github.com/some/other v1.0.0\n",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "go.mod.*")
			assert.Nil(t, err)
			defer os.Remove(f.Name())

			_, err = f.WriteString(tt.content)
			assert.Nil(t, err)
			f.Close()

			result, err := detectGoravelVersionFrom(f.Name())
			if tt.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	t.Run("missing go.mod", func(t *testing.T) {
		_, err := detectGoravelVersionFrom("/nonexistent/path/go.mod")
		assert.NotNil(t, err)
	})
}

func TestSha256sum(t *testing.T) {
	result := sha256sum([]byte("hello"))
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", result)
}

func TestIsSupportedVersion(t *testing.T) {
	tests := []struct {
		version   string
		supported bool
	}{
		{"master", true},
		{"latest", true},
		{"v1.17", true},
		{"v1.18", true},
		{"v2.0", true},
		{"v1.16", false},
		{"v1.0", false},
		{"v0.9", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			assert.Equal(t, tt.supported, isSupportedVersion(tt.version))
		})
	}
}

func TestResolveBranch(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{"v1.17", "v1.17"},
		{"v1.16", "v1.16"},
		{"v1.13", "v1.13"},
		{"v1.99", "v1.99"},
		{"v2.0", "v2.0"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			assert.Equal(t, tt.expected, resolveBranch(tt.version))
		})
	}
}

func TestParseVersionParts(t *testing.T) {
	major, minor := parseVersionParts("v1.17")
	assert.Equal(t, 1, major)
	assert.Equal(t, 17, minor)

	major, minor = parseVersionParts("v2.5")
	assert.Equal(t, 2, major)
	assert.Equal(t, 5, minor)

	major, minor = parseVersionParts("invalid")
	assert.Equal(t, 0, major)
	assert.Equal(t, 0, minor)
}
