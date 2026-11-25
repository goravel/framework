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
			expected:   "APP_ENV=local\nFILESYSTEM_DISK=local",
		},
		{
			name:       "keeps an existing value",
			env:        "FILESYSTEM_DISK=s3",
			envExample: "FILESYSTEM_DISK=s3",
			expected:   "FILESYSTEM_DISK=s3",
		},
		{
			name:       "appends after a commented key",
			env:        "# FILESYSTEM_DISK=s3",
			envExample: "# FILESYSTEM_DISK=s3",
			expected:   "# FILESYSTEM_DISK=s3\nFILESYSTEM_DISK=local",
		},
		{
			name:       "does not duplicate an assignment with spaces",
			env:        "FILESYSTEM_DISK = s3",
			envExample: "FILESYSTEM_DISK = s3",
			expected:   "FILESYSTEM_DISK = s3",
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

			require.NoError(t, install(t, "Storage"))

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
		env      string
		expected string
		absent   bool
	}{
		{
			name:     "writes the driver selector",
			env:      "APP_ENV=local",
			expected: "APP_ENV=local\nFILESYSTEM_DISK=local",
		},
		{
			name:     "overwrites a stale selector",
			env:      "FILESYSTEM_DISK=s3",
			expected: "FILESYSTEM_DISK=local",
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
				writeFile(t, envExamplePath, tt.env)
			}

			require.NoError(t, installDriver(t, "local"))

			if tt.absent {
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)

				require.NoError(t, installDriver(t, "local"))
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)
				return
			}

			assert.Equal(t, tt.expected, readFile(t, envPath))
			assert.Equal(t, tt.expected, readFile(t, envExamplePath))

			env := readFile(t, envPath)
			envExample := readFile(t, envExamplePath)
			require.NoError(t, installDriver(t, "local"))
			assert.Equal(t, env, readFile(t, envPath))
			assert.Equal(t, envExample, readFile(t, envExamplePath))
		})
	}
}

func TestInstallActions_Idempotent(t *testing.T) {
	root := setupProject(t)
	envPath := filepath.Join(root, ".env")
	envExamplePath := filepath.Join(root, ".env.example")
	writeFile(t, envPath, "APP_ENV=local")
	writeFile(t, envExamplePath, "APP_ENV=local")

	require.NoError(t, install(t, "Storage"))
	env := readFile(t, envPath)
	envExample := readFile(t, envExamplePath)

	require.NoError(t, install(t, "Storage"))
	assert.Equal(t, env, readFile(t, envPath))
	assert.Equal(t, envExample, readFile(t, envExamplePath))
	assert.Equal(t, 1, strings.Count(readFile(t, envPath), "FILESYSTEM_DISK"))
}

func TestInstallActions_GeneratesProjectFiles(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	require.NoError(t, install(t, "Storage"))

	assert.Contains(t, readFile(t, filepath.Join(root, "config", "filesystems.go")), `"default": config.Env("FILESYSTEM_DISK", "local"),`)
	assert.FileExists(t, filepath.Join(root, "app", "facades", "storage.go"))

	// Every generated file stays in the temporary project, never in the package directory.
	assert.NoDirExists(t, "config")
	assert.NoDirExists(t, "bootstrap")
	assert.NoDirExists(t, "app")
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

	require.NoError(t, install(t, "Storage"))
	assert.FileExists(t, filepath.Join(root, "config", "filesystems.go"))

	require.NoError(t, uninstall(t, "Storage"))
	assert.NoFileExists(t, filepath.Join(root, "config", "filesystems.go"))
	assert.NoFileExists(t, filepath.Join(root, "app", "facades", "storage.go"))
}

func TestMain_Install(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"setup", "install", "--facade=Storage", "--main-path=goravel"}

	main()

	assert.FileExists(t, filepath.Join(root, "config", "filesystems.go"))
}
