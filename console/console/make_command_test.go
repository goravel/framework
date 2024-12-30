package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestMakeCommand(t *testing.T) {
	makeCommand := &MakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the command name", mock.Anything).Return("", errors.New("the command name cannot be empty")).Once()
	mockContext.EXPECT().Error("the command name cannot be empty").Once()
	assert.Nil(t, makeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("CleanCache").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Console command created successfully").Once()
	assert.Nil(t, makeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/console/commands/clean_cache.go"))

	mockContext.EXPECT().Argument(0).Return("CleanCache").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the command already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, makeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Goravel/CleanCache").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Console command created successfully").Once()
	assert.Nil(t, makeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/console/commands/Goravel/clean_cache.go"))
	assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "package Goravel"))
	assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "type CleanCache struct"))

	assert.Nil(t, file.Remove("app"))
}
