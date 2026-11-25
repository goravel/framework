package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/options"
	"github.com/goravel/framework/support"
)

const appStub = `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().Create()
}
`

// setupProject points the project root at a temporary application skeleton.
func setupProject(t *testing.T) string {
	t.Helper()

	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	writeFile(t, filepath.Join(support.RelativePath, "bootstrap", "app.go"), appStub)

	return support.RelativePath
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(content)
}

// install applies the install actions with the installer options.
func install(t *testing.T, facade string) error {
	t.Helper()

	setup := packages.Setup([]string{"install", "--main-path=goravel", "--facade=" + facade})
	for _, action := range installActions(setup) {
		if err := action.Apply(options.Facade(facade)); err != nil {
			return err
		}
	}

	return nil
}

// installDriver applies the install actions with the installer driver option.
func installDriver(t *testing.T, driver string) error {
	t.Helper()

	setup := packages.Setup([]string{"install", "--main-path=goravel", "--driver=" + driver})
	for _, action := range installActions(setup) {
		if err := action.Apply(options.Driver(driver)); err != nil {
			return err
		}
	}

	return nil
}

func TestInstallActions_Env(t *testing.T) {
	tests := []struct {
		name       string
		env        string
		envExample string
		absent     bool
		expected   string
	}{
		{
			name:       "appends the default",
			env:        "APP_ENV=local",
			envExample: "APP_ENV=local",
			expected:   "APP_ENV=local\nQUEUE_CONNECTION=sync",
		},
		{
			name:       "keeps an existing value",
			env:        "QUEUE_CONNECTION=redis",
			envExample: "QUEUE_CONNECTION=redis",
			expected:   "QUEUE_CONNECTION=redis",
		},
		{
			name:       "appends after a commented key",
			env:        "# QUEUE_CONNECTION=redis",
			envExample: "# QUEUE_CONNECTION=redis",
			expected:   "# QUEUE_CONNECTION=redis\nQUEUE_CONNECTION=sync",
		},
		{
			name:       "does not duplicate an assignment with spaces",
			env:        "QUEUE_CONNECTION = redis",
			envExample: "QUEUE_CONNECTION = redis",
			expected:   "QUEUE_CONNECTION = redis",
		},
		{
			name:   "leaves missing files absent",
			absent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := setupProject(t)
			envPath := filepath.Join(root, ".env")
			envExamplePath := filepath.Join(root, ".env.example")
			if !tt.absent {
				writeFile(t, envPath, tt.env)
				writeFile(t, envExamplePath, tt.envExample)
			}

			require.NoError(t, install(t, "Queue"))

			if tt.absent {
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)
				return
			}

			assert.Equal(t, tt.expected, readFile(t, envPath))
			assert.Equal(t, tt.expected, readFile(t, envExamplePath))
		})
	}
}

func TestInstallActions_Driver(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		env      string
		expected string
		config   bool
		absent   bool
	}{
		{
			name:     "writes the sync selector",
			driver:   "sync",
			env:      "APP_ENV=local",
			expected: "APP_ENV=local\nQUEUE_CONNECTION=sync",
		},
		{
			name:     "overwrites a stale selector",
			driver:   "sync",
			env:      "QUEUE_CONNECTION=redis",
			expected: "QUEUE_CONNECTION=sync",
		},
		{
			name:     "writes the database selector and config",
			driver:   "database",
			env:      "QUEUE_CONNECTION=sync",
			expected: "QUEUE_CONNECTION=database",
			config:   true,
		},
		{
			name:   "leaves missing files absent",
			driver: "sync",
			absent: true,
		},
		{
			name:   "writes no env selector when the files are absent",
			driver: "database",
			absent: true,
			config: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := setupProject(t)
			envPath := filepath.Join(root, ".env")
			envExamplePath := filepath.Join(root, ".env.example")
			configPath := filepath.Join(root, "config", "queue.go")
			if tt.config {
				// Generate config/queue.go from the stub the installer writes.
				require.NoError(t, install(t, "Queue"))
			}
			if !tt.absent {
				writeFile(t, envPath, tt.env)
				writeFile(t, envExamplePath, tt.env)
			}

			require.NoError(t, installDriver(t, tt.driver))

			if tt.absent {
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)
			} else {
				assert.Equal(t, tt.expected, readFile(t, envPath))
				assert.Equal(t, tt.expected, readFile(t, envExamplePath))
			}

			if tt.config {
				content := readFile(t, configPath)
				// The comment above the selector survives the replacement.
				assert.Regexp(t, "(?m)^\\t*// Default Queue Connection Name\n\\t*\"default\": config\\.Env\\(\"QUEUE_CONNECTION\", \"database\"\\),$", content)
				assert.NotContains(t, content, `"default": config.Env("QUEUE_CONNECTION", "sync"),`)
			}

			// Applying the same driver actions twice changes nothing.
			snapshot := map[string]string{}
			for _, path := range []string{envPath, envExamplePath, configPath} {
				if content, err := os.ReadFile(path); err == nil {
					snapshot[path] = string(content)
				}
			}

			require.NoError(t, installDriver(t, tt.driver))

			for path, content := range snapshot {
				assert.Equal(t, content, readFile(t, path))
			}
			if tt.absent {
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)
			}
		})
	}
}

func TestInstallActions_Idempotent(t *testing.T) {
	root := setupProject(t)
	envPath := filepath.Join(root, ".env")
	envExamplePath := filepath.Join(root, ".env.example")
	writeFile(t, envPath, "APP_ENV=local")
	writeFile(t, envExamplePath, "APP_ENV=local")

	require.NoError(t, install(t, "Queue"))
	env := readFile(t, envPath)
	envExample := readFile(t, envExamplePath)

	require.NoError(t, install(t, "Queue"))
	assert.Equal(t, env, readFile(t, envPath))
	assert.Equal(t, envExample, readFile(t, envExamplePath))
	assert.Equal(t, 1, strings.Count(readFile(t, envPath), "QUEUE_CONNECTION"))
}

func TestInstallActions_GeneratesProjectFiles(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	require.NoError(t, install(t, "Queue"))

	assert.Contains(t, readFile(t, filepath.Join(root, "config", "queue.go")), `"default": config.Env("QUEUE_CONNECTION", "sync"),`)
	assert.FileExists(t, filepath.Join(root, "app", "facades", "queue.go"))

	// Every generated file stays in the temporary project, never in the package directory.
	assert.NoDirExists(t, "config")
	assert.NoDirExists(t, "bootstrap")
	assert.NoDirExists(t, "app")
	assert.NoDirExists(t, "database")
	assert.NoFileExists(t, ".env")
	assert.NoFileExists(t, ".env.example")
}

func uninstall(t *testing.T, facade string) error {
	t.Helper()

	setup := packages.Setup([]string{"uninstall", "--main-path=goravel", "--facade=" + facade})
	for _, action := range uninstallActions(setup) {
		if err := action.Apply(options.Facade(facade)); err != nil {
			return err
		}
	}

	return nil
}

func TestUninstallActions_RemovesGeneratedFiles(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	require.NoError(t, install(t, "Queue"))
	assert.FileExists(t, filepath.Join(root, "config", "queue.go"))

	require.NoError(t, uninstall(t, "Queue"))
	assert.NoFileExists(t, filepath.Join(root, "config", "queue.go"))
	assert.NoFileExists(t, filepath.Join(root, "app", "facades", "queue.go"))
}

func TestMain_Install(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"setup", "install", "--facade=Queue", "--main-path=goravel"}

	main()

	assert.FileExists(t, filepath.Join(root, "config", "queue.go"))
}
