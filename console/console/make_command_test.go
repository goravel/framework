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

	makeCommand := &MakeCommand{}

	t.Run("empty name", func(t *testing.T) {
		assert.NoError(t, file.PutContent("app/console/kernel.go", Stubs{}.Kernel()))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("").Once()
		mockContext.EXPECT().Ask("Enter the command name", mock.Anything).Return("", errors.New("the command name cannot be empty")).Once()
		mockContext.EXPECT().Error("the command name cannot be empty").Once()
		assert.Nil(t, makeCommand.Handle(mockContext))
	})

	t.Run("command register failed", func(t *testing.T) {
		assert.NoError(t, file.PutContent("app/console/kernel.go", `package console

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
		mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
			return strings.HasPrefix(msg, "command register failed:")
		})).Once()
		assert.Nil(t, makeCommand.Handle(mockContext))
		assert.True(t, file.Exists("app/console/commands/clean_cache.go"))
		assert.True(t, file.Contain("app/console/commands/clean_cache.go", "app:clean-cache"))
	})

	t.Run("command already exists", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("CleanCache").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Error("the command already exists. Use the --force or -f flag to overwrite").Once()
		assert.Nil(t, makeCommand.Handle(mockContext))
	})

	t.Run("command create and register successfully", func(t *testing.T) {
		assert.NoError(t, file.PutContent("app/console/kernel.go", Stubs{}.Kernel()))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("Goravel/CleanCache").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Console command created successfully").Once()
		mockContext.EXPECT().Success("Console command registered successfully").Once()

		assert.Nil(t, makeCommand.Handle(mockContext))
		assert.True(t, file.Exists("app/console/commands/Goravel/clean_cache.go"))
		assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "package Goravel"))
		assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "type CleanCache struct"))
		assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "app:goravel-clean-cache"))
		assert.True(t, file.Contain("app/console/kernel.go", "app/console/commands/Goravel"))
		assert.True(t, file.Contain("app/console/kernel.go", "&Goravel.CleanCache{}"))
	})
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
