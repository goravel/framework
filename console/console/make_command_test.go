package console

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

var (
	appServiceProvider = `package providers

import (
	"github.com/goravel/framework/contracts/foundation"
)

type AppServiceProvider struct {
}

func (receiver *AppServiceProvider) Register(app foundation.Application) {}

func (receiver *AppServiceProvider) Boot(app foundation.Application) {}
`
)

func TestMakeCommand(t *testing.T) {
	defer func() {
		assert.Nil(t, file.Remove("app"))
	}()

	kernelPath := filepath.Join("app", "console", "kernel.go")
	appServiceProviderPath := filepath.Join("app", "providers", "app_service_provider.go")
	makeCommand := &MakeCommand{}

	t.Run("empty name", func(t *testing.T) {
		assert.NoError(t, file.PutContent(kernelPath, Stubs{}.Kernel()))
		assert.NoError(t, file.PutContent(appServiceProviderPath, appServiceProvider))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("").Once()
		mockContext.EXPECT().Ask("Enter the command name", mock.Anything).Return("", errors.New("the command name cannot be empty")).Once()
		mockContext.EXPECT().Error("the command name cannot be empty").Once()
		assert.Nil(t, makeCommand.Handle(mockContext))
	})

	t.Run("command register failed", func(t *testing.T) {
		assert.NoError(t, file.PutContent(kernelPath, `package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
)

type Kernel struct {
}

func (kernel Kernel) Schedule() []schedule.Event {
	return []schedule.Event{}
}
`))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("CleanCache").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Console command created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "modify go file") && strings.Contains(msg, "kernel.go")
		})).Once()
		assert.Nil(t, makeCommand.Handle(mockContext))

		cleanCachePath := filepath.Join("app", "console", "commands", "clean_cache.go")
		assert.True(t, file.Exists(cleanCachePath))
		assert.True(t, file.Contain(cleanCachePath, "app:clean-cache"))
	})

	t.Run("command already exists", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("CleanCache").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Error("the command already exists. Use the --force or -f flag to overwrite").Once()
		assert.Nil(t, makeCommand.Handle(mockContext))
	})

	t.Run("command create and register successfully", func(t *testing.T) {
		assert.NoError(t, file.PutContent(kernelPath, Stubs{}.Kernel()))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("Goravel/CleanCache").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Console command created successfully").Once()
		mockContext.EXPECT().Success("Console command registered successfully").Once()

		assert.Nil(t, makeCommand.Handle(mockContext))

		cleanCachePath := filepath.Join("app", "console", "commands", "Goravel", "clean_cache.go")
		assert.True(t, file.Exists(cleanCachePath))
		assert.True(t, file.Contain(cleanCachePath, "package Goravel"))
		assert.True(t, file.Contain(cleanCachePath, "type CleanCache struct"))
		assert.True(t, file.Contain(cleanCachePath, "app:goravel-clean-cache"))
		assert.True(t, file.Contain(kernelPath, "app/console/commands/Goravel"))
		assert.True(t, file.Contain(kernelPath, "&Goravel.CleanCache{}"))
	})
}

func TestMakeCommand_AddCommandToBootstrapSetup(t *testing.T) {
	makeCommand := &MakeCommand{}
	bootstrapPath := filepath.Join("bootstrap", "app.go")

	// Ensure clean state before test
	defer func() {
		assert.Nil(t, file.Remove("bootstrap"))
	}()

	// Create bootstrap/app.go with foundation.Setup()
	bootstrapContent := `package bootstrap

import (
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().Run()
}
`
	assert.NoError(t, file.PutContent(bootstrapPath, bootstrapContent))

	// Create mock context
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("CleanCache").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Console command created successfully").Once()
	mockContext.EXPECT().Success("Console command registered successfully").Once()

	// Execute
	err := makeCommand.Handle(mockContext)

	// Assert
	assert.Nil(t, err)

	// Verify the command file was created
	cleanCachePath := filepath.Join("app", "console", "commands", "clean_cache.go")
	assert.True(t, file.Exists(cleanCachePath))
	assert.True(t, file.Contain(cleanCachePath, "app:clean-cache"))

	defer assert.NoError(t, file.Remove(cleanCachePath))

	// Verify bootstrap/app.go was modified with AddCommand
	bootstrapContent, readErr := file.GetContent(bootstrapPath)
	assert.NoError(t, readErr)
	expectedContent := `package bootstrap

import (
	"github.com/goravel/framework/app/console/commands"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation"
)

func Boot() {
	foundation.Setup().
		WithCommands([]console.Command{
			&commands.CleanCache{},
		}).Run()
}
`
	assert.Equal(t, expectedContent, bootstrapContent)
}

func TestMakeCommand_initKernel(t *testing.T) {
	makeCommand := &MakeCommand{}
	kernelPath := filepath.Join("app", "console", "kernel.go")
	appServiceProviderPath := filepath.Join("app", "providers", "app_service_provider.go")

	tests := []struct {
		name   string
		setup  func()
		assert func(err error)
	}{
		{
			name: "happy path",
			setup: func() {
				// Create app_service_provider.go with basic structure for modification
				assert.NoError(t, file.PutContent(appServiceProviderPath, appServiceProvider))
			},
			assert: func(err error) {
				assert.NoError(t, err)

				kernel, err := file.GetContent(kernelPath)
				assert.NoError(t, err)
				assert.Equal(t, Stubs{}.Kernel(), kernel)

				appServiceProvider, err := file.GetContent(appServiceProviderPath)
				assert.NoError(t, err)
				assert.True(t, strings.Contains(appServiceProvider, "github.com/goravel/framework/contracts/foundation"))
				assert.True(t, strings.Contains(appServiceProvider, "github.com/goravel/framework/app/facades"))
				assert.True(t, strings.Contains(appServiceProvider, "github.com/goravel/framework/app/console"))
				assert.True(t, strings.Contains(appServiceProvider, "facades.Artisan().Register(console.Kernel{}.Commands())"))
			},
		},
		{
			name: "kernel file already exists, modify app_service_provider.go successfully",
			setup: func() {
				// Create the kernel file
				assert.NoError(t, file.PutContent(kernelPath, Stubs{}.Kernel()))

				// Create app_service_provider.go with basic structure for modification
				assert.NoError(t, file.PutContent(appServiceProviderPath, appServiceProvider))
			},
			assert: func(err error) {
				assert.NoError(t, err)

				appServiceProvider, err := file.GetContent(appServiceProviderPath)
				assert.NoError(t, err)
				assert.True(t, strings.Contains(appServiceProvider, "github.com/goravel/framework/contracts/foundation"))
				assert.True(t, strings.Contains(appServiceProvider, "github.com/goravel/framework/app/facades"))
				assert.True(t, strings.Contains(appServiceProvider, "github.com/goravel/framework/app/console"))
				assert.True(t, strings.Contains(appServiceProvider, "facades.Artisan().Register(console.Kernel{}.Commands())"))
			},
		},
		{
			name: "fail to modify app_service_provider.go",
			setup: func() {
				// Create app_service_provider.go with basic structure for modification
				appServiceProvider := `package providers

import (
	"github.com/goravel/framework/contracts/foundation"
)

type AppServiceProvider struct {}

func (receiver *AppServiceProvider) Boot(app foundation.Application) {}
`
				assert.NoError(t, file.PutContent(appServiceProviderPath, appServiceProvider))
			},
			assert: func(err error) {
				if env.IsWindows() {
					assert.Equal(t, "modify go file 'app\\providers\\app_service_provider.go' failed: 1 out of 1 matchers did not match", err.Error())
				} else {
					assert.Equal(t, "modify go file 'app/providers/app_service_provider.go' failed: 1 out of 1 matchers did not match", err.Error())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := makeCommand.initKernel()

			tt.assert(err)
			assert.Nil(t, file.Remove("app"))
		})
	}
}
