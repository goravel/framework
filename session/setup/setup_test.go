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
			name:       "appends the defaults",
			env:        "APP_ENV=local",
			envExample: "APP_ENV=local",
			expected:   "APP_ENV=local\nSESSION_DRIVER=file\nSESSION_LIFETIME=120",
		},
		{
			name:       "keeps existing values",
			env:        "SESSION_DRIVER=redis\nSESSION_LIFETIME=60",
			envExample: "SESSION_DRIVER=redis\nSESSION_LIFETIME=60",
			expected:   "SESSION_DRIVER=redis\nSESSION_LIFETIME=60",
		},
		{
			name:       "appends after a commented key",
			env:        "# SESSION_DRIVER=redis\nSESSION_LIFETIME=60",
			envExample: "# SESSION_DRIVER=redis\nSESSION_LIFETIME=60",
			expected:   "# SESSION_DRIVER=redis\nSESSION_LIFETIME=60\nSESSION_DRIVER=file",
		},
		{
			name:       "does not duplicate an assignment with spaces",
			env:        "SESSION_DRIVER = redis\nSESSION_LIFETIME = 60",
			envExample: "SESSION_DRIVER = redis\nSESSION_LIFETIME = 60",
			expected:   "SESSION_DRIVER = redis\nSESSION_LIFETIME = 60",
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

			require.NoError(t, install(t, "Session"))

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
			expected: "APP_ENV=local\nSESSION_DRIVER=file",
		},
		{
			name:     "overwrites a stale selector",
			env:      "SESSION_DRIVER=redis",
			expected: "SESSION_DRIVER=file",
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

			require.NoError(t, installDriver(t, "file"))

			if tt.absent {
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)

				require.NoError(t, installDriver(t, "file"))
				assert.NoFileExists(t, envPath)
				assert.NoFileExists(t, envExamplePath)
				return
			}

			assert.Equal(t, tt.expected, readFile(t, envPath))
			assert.Equal(t, tt.expected, readFile(t, envExamplePath))

			env := readFile(t, envPath)
			envExample := readFile(t, envExamplePath)
			require.NoError(t, installDriver(t, "file"))
			assert.Equal(t, env, readFile(t, envPath))
			assert.Equal(t, envExample, readFile(t, envExamplePath))
		})
	}
}

func TestInstallActions_KeepsSessionLifetime(t *testing.T) {
	root := setupProject(t)
	envPath := filepath.Join(root, ".env")
	envExamplePath := filepath.Join(root, ".env.example")
	writeFile(t, envPath, "SESSION_LIFETIME=60")
	writeFile(t, envExamplePath, "SESSION_LIFETIME=60")

	require.NoError(t, install(t, "Session"))

	for _, path := range []string{envPath, envExamplePath} {
		content := readFile(t, path)
		assert.Equal(t, 1, strings.Count(content, "SESSION_LIFETIME"))
		assert.Equal(t, "SESSION_LIFETIME=60\nSESSION_DRIVER=file", content)
	}
}

func TestInstallActions_Idempotent(t *testing.T) {
	root := setupProject(t)
	envPath := filepath.Join(root, ".env")
	envExamplePath := filepath.Join(root, ".env.example")
	writeFile(t, envPath, "APP_ENV=local")
	writeFile(t, envExamplePath, "APP_ENV=local")

	require.NoError(t, install(t, "Session"))
	env := readFile(t, envPath)
	envExample := readFile(t, envExamplePath)

	require.NoError(t, install(t, "Session"))
	assert.Equal(t, env, readFile(t, envPath))
	assert.Equal(t, envExample, readFile(t, envExamplePath))
	assert.Equal(t, 1, strings.Count(readFile(t, envPath), "SESSION_DRIVER"))
	assert.Equal(t, 1, strings.Count(readFile(t, envPath), "SESSION_LIFETIME"))
}

func TestInstallActions_GeneratesProjectFiles(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	require.NoError(t, install(t, "Session"))

	assert.Contains(t, readFile(t, filepath.Join(root, "config", "session.go")), `"default": config.Env("SESSION_DRIVER", "file"),`)
	assert.FileExists(t, filepath.Join(root, "app", "facades", "session.go"))

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

	require.NoError(t, install(t, "Session"))
	assert.FileExists(t, filepath.Join(root, "config", "session.go"))

	require.NoError(t, uninstall(t, "Session"))
	assert.NoFileExists(t, filepath.Join(root, "config", "session.go"))
	assert.NoFileExists(t, filepath.Join(root, "app", "facades", "session.go"))
}

func TestMain_Install(t *testing.T) {
	root := setupProject(t)
	writeFile(t, filepath.Join(root, ".env"), "APP_ENV=local")
	writeFile(t, filepath.Join(root, ".env.example"), "APP_ENV=local")

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"setup", "install", "--facade=Session", "--main-path=goravel"}

	main()

	assert.FileExists(t, filepath.Join(root, "config", "session.go"))
}
