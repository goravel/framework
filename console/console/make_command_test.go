package console

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

func TestMakeCommand(t *testing.T) {
	makeCommand := &MakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the command name", mock.Anything).Return("", errors.New("the command name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, makeCommand.Handle(mockContext))
	}), "the command name cannot be empty")

	mockContext.On("Argument", 0).Return("CleanCache").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, makeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/console/commands/clean_cache.go"))

	mockContext.On("Argument", 0).Return("CleanCache").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, makeCommand.Handle(mockContext))
	}), "the command already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("Goravel/CleanCache").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, makeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/console/commands/Goravel/clean_cache.go"))
	assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "package Goravel"))
	assert.True(t, file.Contain("app/console/commands/Goravel/clean_cache.go", "type CleanCache struct"))

	assert.Nil(t, file.Remove("app"))
}
