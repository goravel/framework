package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support"
)

func TestMainPath(t *testing.T) {
	assert.Equal(t, "github.com/goravel/framework", MainPath())
}

func TestIsBootstrapSetup(t *testing.T) {
	originalBootstrap := support.Config.Paths.Bootstrap
	originalRelativePath := support.RelativePath
	t.Cleanup(func() {
		support.Config.Paths.Bootstrap = originalBootstrap
		support.RelativePath = originalRelativePath
	})

	tests := []struct {
		name      string
		setup     func(dir string)
		bootstrap string
		expect    bool
	}{
		{
			name:      "empty bootstrap path returns false",
			bootstrap: "",
			setup:     func(dir string) {},
			expect:    false,
		},
		{
			name:      "bootstrap file does not exist returns false",
			bootstrap: "bootstrap",
			setup:     func(dir string) {},
			expect:    false,
		},
		{
			name:      "bootstrap app.go exists but missing setup call returns false",
			bootstrap: "bootstrap",
			setup: func(dir string) {
				bootstrapDir := filepath.Join(dir, "bootstrap")
				assert.NoError(t, os.MkdirAll(bootstrapDir, 0755))
				assert.NoError(t, os.WriteFile(filepath.Join(bootstrapDir, "app.go"), []byte(`package bootstrap`), 0644))
			},
			expect: false,
		},
		{
			name:      "bootstrap app.go contains foundation.Setup() returns true",
			bootstrap: "bootstrap",
			setup: func(dir string) {
				bootstrapDir := filepath.Join(dir, "bootstrap")
				assert.NoError(t, os.MkdirAll(bootstrapDir, 0755))
				content := `package bootstrap

import "github.com/goravel/framework/foundation"

var App = foundation.Setup().Create()
`
				assert.NoError(t, os.WriteFile(filepath.Join(bootstrapDir, "app.go"), []byte(content), 0644))
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)
			support.Config.Paths.Bootstrap = tt.bootstrap
			support.RelativePath = dir

			assert.Equal(t, tt.expect, IsBootstrapSetup())
		})
	}
}
