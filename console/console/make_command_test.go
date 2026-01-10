package console

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

var (
	kernel = `package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
)

type Kernel struct {
}

func (kernel Kernel) Commands() []console.Command {
	return []console.Command{}
}

func (kernel Kernel) Schedule() []schedule.Event {
	return []schedule.Event{}
}
`
)

func TestMakeCommand(t *testing.T) {
	defer func() {
		assert.Nil(t, file.Remove("app"))
	}()

	kernelPath := filepath.Join("app", "console", "kernel.go")
	makeCommand := &MakeCommand{}

	t.Run("empty name", func(t *testing.T) {
		assert.NoError(t, file.PutContent(kernelPath, kernel))

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
		assert.NoError(t, file.PutContent(kernelPath, kernel))

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
	commandsPath := filepath.Join("bootstrap", "commands.go")

	// Ensure clean state before test
	defer func() {
		assert.Nil(t, file.Remove("bootstrap"))
	}()

	// Create bootstrap/app.go with foundation.Setup()
	bootstrapContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
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
	expectedAppContent := `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).Start()
}
`
	assert.Equal(t, expectedAppContent, bootstrapContent)

	// Verify bootstrap/commands.go was created
	commandsContent, readErr := file.GetContent(commandsPath)
	assert.NoError(t, readErr)
	expectedCommandsContent := `package bootstrap

import (
	"github.com/goravel/framework/app/console/commands"
	"github.com/goravel/framework/contracts/console"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.CleanCache{},
	}
}
`
	assert.Equal(t, expectedCommandsContent, commandsContent)
}
